package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/model"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"confkeeper/utils/captcha"
	"confkeeper/utils/config"
	"confkeeper/utils/ldap_client"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginReq struct {
	Username   string `json:"username" binding:"required,min=1,max=255"`
	Password   string `json:"password" binding:"required,min=1,max=255"`
	CaptchaID  string `json:"captcha_id" binding:"required,min=1,max=255"`
	Captcha    string `json:"captcha" binding:"required,min=1,max=10"`
	RememberMe bool   `json:"remember_me" binding:"omitempty"`
}

type LoginData struct {
	Token string `json:"token"`
}

type LoginResp struct {
	Code response.Code `json:"code"`
	Msg  string        `json:"msg"`
	Data *LoginData    `json:"data"`
}

// UserLogin 用户登录
//
//	@Tags			用户
//	@Summary		用户登录
//	@Description	用户登录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		LoginReq	true	"登录凭证"
//	@Success		200	{object}	LoginResp
//	@router			/api/user/login [POST]
func UserLogin(c *gin.Context) {
	req := new(LoginReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(LoginResp)

	// 验证验证码
	if !captcha.Store.Verify(req.CaptchaID, req.Captcha, true) {
		c.JSON(http.StatusOK, &LoginResp{
			Code: response.Code_CaptchaErr,
			Msg:  "验证码错误或已过期",
		})
		return
	}

	userData, err := dal.UserLogin(req.Username)
	if err != nil {
		// 如果数据库没有用户，尝试LDAP登录
		if config.Cfg.Ldap.Enabled {
			success, _, ldapErr := ldap_client.LDAPAuth(req.Username, req.Password)
			if success {
				// LDAP登录成功，注册用户到数据库
				userData = &model.User{
					Username: req.Username,
					Password: utils.MD5(req.Password),
					Enable:   true,
				}
				if createErr := dal.CreateUser([]*model.User{userData}); createErr != nil {
					c.JSON(http.StatusOK, &LoginResp{Code: response.Code_DBErr, Msg: "用户同步失败: " + createErr.Error()})
					return
				}
				// 注册成功后继续走登录逻辑(生成token)
			} else {
				// LDAP登录失败
				msg := "用户不存在"
				if ldapErr != nil {
					msg += " 或 " + ldapErr.Error()
				}
				c.JSON(http.StatusOK, &LoginResp{Code: response.Code_DBErr, Msg: msg})
				return
			}
		} else {
			c.JSON(http.StatusOK, &LoginResp{Code: response.Code_DBErr, Msg: "用户不存在或密码错误"})
			return
		}
	} else {
		// 数据库有用户，验证密码
		if userData.Password != utils.MD5(req.Password) {
			c.JSON(http.StatusOK, &LoginResp{Code: response.Code_PasswordErr, Msg: "密码错误"})
			return
		}
	}

	var token string
	if req.RememberMe {
		token, _ = utils.GenerateToken(userData.ID, req.Username)
	} else {
		//如果没有选记住我就1小时token
		token, _ = utils.GenerateToken(userData.ID, req.Username, 60)
	}

	resp.Code = response.Code_Success
	resp.Msg = "登录成功"
	resp.Data = &LoginData{
		Token: token,
	}
	c.JSON(http.StatusOK, resp)
}
