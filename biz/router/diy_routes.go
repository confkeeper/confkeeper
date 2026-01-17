package router

import (
	"confkeeper/biz/handler"

	"github.com/gin-gonic/gin"
)

func diyRoutes(apiGroup *gin.RouterGroup) {
	apiGroup.GET("/ping", handler.Ping)
	apiGroup.GET("/server_info", handler.ServerInfo)
	apiGroup.GET("/is_demo", handler.GetDemo)
	apiGroup.GET("/metrics", handler.Metrics)
}
