package permission

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"confkeeper/utils/config"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	Role     string `json:"role" binding:"required,min=1,max=255"`
	Resource string `json:"resource" binding:"required,min=1,max=255"`
	Action   string `json:"action" binding:"required,min=1,max=255"`
}

// CreatePermission 创建权限
//
//	@Tags			权限管理
//	@Summary		创建权限
//	@Description	为角色分配权限
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		CreateReq	true	"权限信息"
//	@Success		200	{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/permission/add [PUT]
func CreatePermission(c *gin.Context) {
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

	// 检查角色是否存在
	roleExist, err := dal.IsRoleExistsInRoles(req.Role)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查角色是否存在失败: " + err.Error(),
		})
		return
	}
	if !roleExist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Err,
			Msg:  "角色不存在",
		})
		return
	}

	// 检查权限是否存在
	if !slices.Contains(config.Cfg.Confkeeper.ActionType, req.Action) {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Unauthorized,
			Msg:  "没有这个权限",
		})
		return
	}

	// 检查命名空间是否已存在
	exist, err := dal.IsTenantIdExists(req.Resource)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查命名空间失败: " + err.Error(),
		})
		return
	}
	if !exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_AlreadyExists,
			Msg:  "该命名空间不存在",
		})
		return
	}

	// 检查权限是否已存在
	exist, err = dal.IsPermissionExists(req.Role, req.Resource, req.Action)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查权限失败: " + err.Error(),
		})
		return
	}
	if exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_AlreadyExists,
			Msg:  "该权限已存在",
		})
		return
	}

	// 创建权限
	if err = dal.AddRolePermission(req.Role, req.Resource, req.Action); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_DBErr, Msg: "权限创建失败: " + err.Error()})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "创建权限成功"

	c.JSON(http.StatusOK, resp)
}
