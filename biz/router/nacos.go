package router

import (
	hConfigInfo "confkeeper/biz/handler/config_info"
	hUser "confkeeper/biz/handler/user"

	"github.com/gin-gonic/gin"
)

func nacosRoutes(r *gin.Engine) {
	nacosGroup := r.Group("/nacos/v1")
	{
		nacosGroup.POST("/cs/configs", hConfigInfo.NacosUpdateConfig)
		nacosGroup.GET("/cs/configs", hConfigInfo.NacosGetConfig)
		nacosGroup.POST("/auth/login", hUser.NacosUserLogin)
	}
}
