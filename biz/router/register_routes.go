package router

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	apiGroup := r.Group("/api")
	diyRoutes(apiGroup)
	configInfoRoutes(apiGroup)
	permissionRoutes(apiGroup)
	roleRoutes(apiGroup)
	tenantRoutes(apiGroup)
	userRoutes(apiGroup)
	nacosRoutes(r)
}
