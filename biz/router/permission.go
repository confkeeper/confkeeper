package router

import (
	hPermission "confkeeper/biz/handler/permission"
	"confkeeper/biz/mw"
	"confkeeper/utils"

	"github.com/gin-gonic/gin"
)

func permissionRoutes(apiGroup *gin.RouterGroup) {
	permissionGroup := apiGroup.Group("/permission")
	permissionGroup.Use(mw.JWTAuthMiddleware())
	{
		permissionGroup.PUT("/add", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hPermission.CreatePermission)
		permissionGroup.DELETE("/delete", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hPermission.DeletePermission)
		permissionGroup.GET("/list", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hPermission.PermissionList)
	}

}
