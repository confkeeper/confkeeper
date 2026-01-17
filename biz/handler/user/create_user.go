package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/model"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	Username string `json:"username" binding:"required,min=1,max=255"`
}

// CreateUser 创建用户
//
//	@Tags			用户
//	@Summary		创建用户
//	@Description	创建用户
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		CreateReq	true	"用户信息"
//	@Success		200	{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/user/add [PUT]
func CreateUser(c *gin.Context) {
	req := new(CreateReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	// 先检查用户名是否已存在
	exist, err := dal.IsUsernameExists(req.Username)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查用户名失败: " + err.Error(),
		})
		return
	}
	if exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_AlreadyExists,
			Msg:  "该用户已存在",
		})
		return
	}

	err = utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Unauthorized,
			Msg:  err.Error(),
		})
		return
	}

	u := &model.User{
		Username: req.Username,
		Password: "",
		Enable:   true,
	}

	if err = dal.CreateUser([]*model.User{u}); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_DBErr, Msg: "用户新建失败: " + err.Error()})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "新建用户成功"

	c.JSON(http.StatusOK, resp)
}
