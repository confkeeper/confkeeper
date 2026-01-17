package router

import (
	hTenant "confkeeper/biz/handler/tenant"
	"confkeeper/biz/mw"

	"github.com/gin-gonic/gin"
)

func tenantRoutes(apiGroup *gin.RouterGroup) {
	tenantGroup := apiGroup.Group("/tenant")
	tenantGroup.Use(mw.JWTAuthMiddleware())
	{
		tenantGroup.PUT("/add", hTenant.CreateTenant)
		tenantGroup.DELETE("/delete/:id", hTenant.DeleteTenant)
		tenantGroup.GET("/list", hTenant.TenantList)
	}
}
