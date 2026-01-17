package router

import (
	hRole "confkeeper/biz/handler/role"
	"confkeeper/biz/mw"

	"github.com/gin-gonic/gin"
)

func roleRoutes(apiGroup *gin.RouterGroup) {
	roleGroup := apiGroup.Group("/role")
	roleGroup.Use(mw.JWTAuthMiddleware())
	{
		roleGroup.PUT("/add", hRole.CreateRole)
		roleGroup.DELETE("/delete/:role", hRole.DeleteRole)
		roleGroup.GET("/list", hRole.RoleList)
	}
}
