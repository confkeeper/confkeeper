package router

import (
	hUser "confkeeper/biz/handler/user"
	"confkeeper/biz/mw"
	"confkeeper/utils"

	"github.com/gin-gonic/gin"
)

func userRoutes(apiGroup *gin.RouterGroup) {
	userGroup := apiGroup.Group("/user")
	{
		userGroup.PUT("/add", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hUser.CreateUser)
		userGroup.DELETE("/delete/:user_id", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hUser.DeleteUser)
		userGroup.POST("/update/:user_id", mw.JWTAuthMiddleware(), hUser.UpdateUser)
		userGroup.POST("/change_passwd/:user_id", mw.JWTAuthMiddleware(), hUser.ChangePasswd)
		userGroup.GET("/captcha", hUser.GenerateCaptcha)
		userGroup.POST("/login", hUser.UserLogin)
		userGroup.POST("/logout", mw.JWTAuthMiddleware(), hUser.UserLogout)
		userGroup.GET("/list", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hUser.UserList)
		userGroup.GET("/info/:user_id", mw.JWTAuthMiddleware(), hUser.UserInfo)
		userGroup.POST("/sync-ldap", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hUser.SyncLDAPUsers)
	}
}
