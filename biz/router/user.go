package router

import (
	hUser "confkeeper/biz/handler/user"
	"confkeeper/biz/mw"

	"github.com/gin-gonic/gin"
)

func userRoutes(apiGroup *gin.RouterGroup) {
	userGroup := apiGroup.Group("/user")
	{
		userGroup.PUT("/add", mw.JWTAuthMiddleware(), hUser.CreateUser)
		userGroup.DELETE("/delete/:user_id", mw.JWTAuthMiddleware(), hUser.DeleteUser)
		userGroup.POST("/update/:user_id", mw.JWTAuthMiddleware(), hUser.UpdateUser)
		userGroup.POST("/change_passwd/:user_id", mw.JWTAuthMiddleware(), hUser.ChangePasswd)
		userGroup.GET("/captcha", hUser.GenerateCaptcha)
		userGroup.POST("/login", hUser.UserLogin)
		userGroup.GET("/list", mw.JWTAuthMiddleware(), hUser.UserList)
		userGroup.GET("/info/:user_id", mw.JWTAuthMiddleware(), hUser.UserInfo)
	}
}
