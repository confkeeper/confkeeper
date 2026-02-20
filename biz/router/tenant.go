package router

import (
	hTenant "confkeeper/biz/handler/tenant"
	"confkeeper/biz/mw"
	"confkeeper/utils"

	"github.com/gin-gonic/gin"
)

func tenantRoutes(apiGroup *gin.RouterGroup) {
	tenantGroup := apiGroup.Group("/tenant")
	tenantGroup.Use(mw.JWTAuthMiddleware())
	{
		tenantGroup.PUT("/add", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hTenant.CreateTenant)
		tenantGroup.DELETE("/delete/:id", mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), hTenant.DeleteTenant)
		tenantGroup.GET("/list", mw.JWTAuthMiddleware(), hTenant.TenantList)
	}
}
