package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/utils"
	"confkeeper/utils/captcha"
	"confkeeper/utils/config"
	"confkeeper/utils/ldap_client"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
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
	handler.CommonResp
	Data *LoginData `json:"data"`
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
			CommonResp: handler.CommonResp{
				Code: handler.Code_CaptchaErr,
				Msg:  "验证码错误或已过期",
			},
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
				hashedPassword, err := utils.HashPassword(req.Password)
				if err != nil {
c.JSON(http.StatusInternalServerError, &LoginResp{
						CommonResp: handler.CommonResp{
							Code: handler.Code_DBErr,
							Msg:  "密码加密失败: " + err.Error(),
						},
					})
					return
				}
				userData = &model.User{
					Username: req.Username,
					Password: hashedPassword,
					Enable:   true,
				}
				if createErr := dal.CreateUser([]*model.User{userData}); createErr != nil {
					c.JSON(http.StatusOK, &LoginResp{
						CommonResp: handler.CommonResp{
							Code: handler.Code_DBErr,
							Msg:  "用户同步失败: " + createErr.Error(),
						},
					})
					return
				}
				// 注册成功后继续走登录逻辑(生成token)
			} else {
				// LDAP登录失败
				msg := "用户不存在"
				if ldapErr != nil {
					msg += " 或 " + ldapErr.Error()
				}
				c.JSON(http.StatusOK, &LoginResp{
					CommonResp: handler.CommonResp{
						Code: handler.Code_DBErr,
						Msg:  msg,
					},
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, &LoginResp{
				CommonResp: handler.CommonResp{
					Code: handler.Code_DBErr,
					Msg:  "用户不存在或密码错误",
				},
			})
			return
		}
	} else {
		// 数据库有用户，检查密码是否为空
		if userData.Password == "" {
			// 密码为空，尝试LDAP登录
			if config.Cfg.Ldap.Enabled {
				success, _, ldapErr := ldap_client.LDAPAuth(req.Username, req.Password)
				if success {
					// LDAP登录成功，更新用户密码
					hashedPassword, err := utils.HashPassword(req.Password)
					if err != nil {
						c.JSON(http.StatusOK, &LoginResp{
							CommonResp: handler.CommonResp{
								Code: handler.Code_DBErr,
								Msg:  "密码加密失败: " + err.Error(),
							},
						})
						return
					}
					userData.Password = hashedPassword
					if updateErr := dal.UpdateUser(userData); updateErr != nil {
						c.JSON(http.StatusOK, &LoginResp{
							CommonResp: handler.CommonResp{
								Code: handler.Code_DBErr,
								Msg:  "密码更新失败",
							},
						})
						return
					}
					// 更新成功后继续走登录逻辑(生成token)
				} else {
					// LDAP登录失败
					if ldapErr != nil {
						slog.Errorf("LDAP login failed: %v", ldapErr)
					}
					c.JSON(http.StatusOK, &LoginResp{
						CommonResp: handler.CommonResp{
							Code: handler.Code_PasswordErr,
							Msg:  "密码错误",
						},
					})
					return
				}
			} else {
				c.JSON(http.StatusOK, &LoginResp{
					CommonResp: handler.CommonResp{
						Code: handler.Code_PasswordErr,
						Msg:  "密码错误",
					},
				})
				return
			}
		} else {
			// 密码不为空，验证密码
			if !utils.CheckPasswordHash(req.Password, userData.Password) {
				c.JSON(http.StatusOK, &LoginResp{
					CommonResp: handler.CommonResp{
						Code: handler.Code_PasswordErr,
						Msg:  "密码错误",
					},
				})
				return
			}
		}
	}

	var token string
	if req.RememberMe {
		token, _ = utils.GenerateToken(userData.ID, req.Username)
	} else {
		//如果没有选记住我就1小时token
		token, _ = utils.GenerateToken(userData.ID, req.Username, 60)
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "登录成功",
	}
	resp.Data = &LoginData{
		Token: token,
	}
	c.JSON(http.StatusOK, resp)
}
