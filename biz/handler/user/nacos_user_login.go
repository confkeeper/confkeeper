package user

import (
	"confkeeper/biz/dal"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NacosLoginReq struct {
	Username string `form:"username" binding:"required,min=1,max=255"`
	Password string `form:"password" binding:"required,min=1,max=255"`
}

type NacosLoginResp struct {
	AccessToken string `json:"accessToken"`
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
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(NacosLoginResp)

	userData, err := dal.UserLogin(req.Username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"msg": err.Error()})
		return
	}

	if userData.Password != utils.MD5(req.Password) {
		c.JSON(http.StatusOK, gin.H{"msg": "密码错误"})
		return
	}

	var token string
	token, _ = utils.GenerateToken(userData.ID, req.Username, 1)

	resp.AccessToken = token

	c.JSON(http.StatusOK, resp)
}
