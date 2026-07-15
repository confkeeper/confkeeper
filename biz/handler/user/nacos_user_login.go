package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NacosLoginReq struct {
	Username string `form:"username" binding:"required,min=1,max=255"`
	Password string `form:"password" binding:"required,min=1,max=255"`
}

type NacosLoginData struct {
	AccessToken string `json:"accessToken"`
}

type NacosLoginResp struct {
	handler.CommonResp
	Data *NacosLoginData `json:"data"`
}

// NacosUserLogin 用户登录(nacos兼容)
//
//	@Tags			用户
//	@Tags			nacos兼容
//	@Summary		用户登录(nacos兼容)
//	@Description	用户登录(nacos兼容)
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		NacosLoginReq	true	"登录凭证"
//	@Success		200	{object}	NacosLoginResp
//	@router			/nacos/v1/auth/login [POST]
func NacosUserLogin(c *gin.Context) {
	req := new(NacosLoginReq)
	if err := c.ShouldBind(req); err != nil {
		handler.ParamError(c, err)
		return
	}
	resp := new(NacosLoginResp)

	userData, err := dal.UserLogin(req.Username)
	if err != nil {
		handler.JSON(c, http.StatusOK, handler.Code_DBErr, err.Error())
		return
	}

	if !utils.CheckPasswordHash(req.Password, userData.Password) {
		handler.JSON(c, http.StatusOK, handler.Code_PasswordErr, "密码错误")
		return
	}

	var token string
	token, _ = utils.GenerateToken(userData.ID, req.Username, 1)

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "登录成功",
	}
	resp.Data = &NacosLoginData{
		AccessToken: token,
	}

	c.JSON(http.StatusOK, resp)
}
