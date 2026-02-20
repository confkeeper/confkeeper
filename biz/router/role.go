package router

import (
	hRole "confkeeper/biz/handler/role"
	"confkeeper/biz/mw"
	"confkeeper/utils"

	"github.com/gin-gonic/gin"
)

func roleRoutes(apiGroup *gin.RouterGroup) {
	roleGroup := apiGroup.Group("/role")
	roleGroup.Use(mw.JWTAuthMiddleware())
	{
		roleGroup.PUT("/add", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hRole.CreateRole)
		roleGroup.PUT("/edit", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hRole.EditRole)
		roleGroup.DELETE("/delete/:role", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hRole.DeleteRole)
		roleGroup.GET("/list", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hRole.RoleList)
	}
}
