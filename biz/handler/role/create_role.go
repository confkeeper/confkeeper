package role

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/model"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	Role     string `json:"role" binding:"required,min=1,max=255"`
	Username string `json:"username" binding:"required,min=1,max=255"`
}

// CreateRole 创建角色
//
//	@Tags			角色管理
//	@Summary		创建角色
//	@Description	创建新的角色
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		CreateReq	true	"角色信息"
//	@Success		200	{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/role/add [PUT]
func CreateRole(c *gin.Context) {
	req := new(CreateReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	// 检查是否为管理员
	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Unauthorized,
			Msg:  err.Error(),
		})
		return
	}

	// 先检查用户名是否已存在
	exist, err := dal.IsUsernameExists(req.Username)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查用户名失败: " + err.Error(),
		})
		return
	}
	if !exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_AlreadyExists,
			Msg:  "该用户不存在",
		})
		return
	}

	// 检查角色是否存在
	roleExist, err := dal.IsRoleExistsInRoles(req.Role)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查角色是否存在失败: " + err.Error(),
		})
		return
	}
	if roleExist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Err,
			Msg:  "角色已存在",
		})
		return
	}

	r := &model.Roles{
		Username: req.Username,
		Role:     req.Role,
	}

	if err = dal.CreateRole([]*model.Roles{r}); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_DBErr, Msg: "角色创建失败: " + err.Error()})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "创建角色成功"

	c.JSON(http.StatusOK, resp)
}
