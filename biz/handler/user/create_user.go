package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	Username string `json:"username" binding:"required,min=1,max=255"`
	Password string `json:"password" binding:"required,min=6,max=255"`
}

type CreateUserResp struct {
	handler.CommonResp
}

// CreateUser 创建用户
//
//	@Tags			用户
//	@Summary		创建用户
//	@Description	创建用户
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		CreateReq	true	"用户信息"
//	@Success		200	{object}	CreateUserResp
//	@Security		ApiKeyAuth
//	@router			/api/user/add [PUT]
func CreateUser(c *gin.Context) {
	req := new(CreateReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(CreateUserResp)

	// 先检查用户名是否已存在
	exist, err := dal.IsUsernameExists(req.Username)
	if err != nil {
		c.JSON(http.StatusOK, &CreateUserResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "检查用户名失败: " + err.Error(),
			},
		})
		return
	}
	if exist {
		c.JSON(http.StatusOK, &CreateUserResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_AlreadyExists,
				Msg:  "该用户已存在",
			},
		})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &CreateUserResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "密码加密失败: " + err.Error(),
			},
		})
		return
	}
	u := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Enable:   true,
	}

	if err = dal.CreateUser([]*model.User{u}); err != nil {
		c.JSON(http.StatusOK, &CreateUserResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "用户新建失败: " + err.Error(),
			},
		})
		return
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "新建用户成功",
	}

	c.JSON(http.StatusOK, resp)
}
