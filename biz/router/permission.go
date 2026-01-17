package router

import (
	hPermission "confkeeper/biz/handler/permission"
	"confkeeper/biz/mw"

	"github.com/gin-gonic/gin"
)

func permissionRoutes(apiGroup *gin.RouterGroup) {
	permissionGroup := apiGroup.Group("/permission")
	permissionGroup.Use(mw.JWTAuthMiddleware())
	{
		permissionGroup.PUT("/add", hPermission.CreatePermission)
		permissionGroup.DELETE("/delete", hPermission.DeletePermission)
		permissionGroup.GET("/list", hPermission.PermissionList)
	}

}
