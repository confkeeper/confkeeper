package router

import (
	hConfigInfo "confkeeper/biz/handler/config_info"
	"confkeeper/biz/mw"

	"github.com/gin-gonic/gin"
)

func configInfoRoutes(apiGroup *gin.RouterGroup) {
	configGroup := apiGroup.Group("/config")
	{
		configGroup.PUT("/add", mw.JWTAuthMiddleware(), hConfigInfo.CreateConfig)
		configGroup.DELETE("/delete/:config_id", mw.JWTAuthMiddleware(), hConfigInfo.DeleteConfig)
		configGroup.DELETE("/config/batch_delete ", mw.JWTAuthMiddleware(), hConfigInfo.BatchDeleteConfig)
		configGroup.POST("/update/:config_id", mw.JWTAuthMiddleware(), hConfigInfo.UpdateConfig)
		configGroup.POST("/update_by_file", mw.JWTAuthMiddleware(true), hConfigInfo.UpdateConfigByFile)
		configGroup.POST("/update_by_user", hConfigInfo.UpdateConfigByUser)
		configGroup.GET("/list", mw.JWTAuthMiddleware(), hConfigInfo.ConfigList)
		configGroup.GET("/get/:config_id", mw.JWTAuthMiddleware(), hConfigInfo.ConfigContent)
		configGroup.GET("/get", mw.JWTAuthMiddleware(), hConfigInfo.ConfigContentByParams)
		configGroup.GET("/get_by_file", mw.JWTAuthMiddleware(true), hConfigInfo.GetConfigByFile)
		configGroup.GET("/get_by_user", hConfigInfo.GetConfigByUser)
		configGroup.GET("/get_version/:config_id", mw.JWTAuthMiddleware(), hConfigInfo.ConfigVersion)
		configGroup.POST("/clone", mw.JWTAuthMiddleware(), hConfigInfo.ConfigClone)
		configGroup.POST("/cleanup", mw.JWTAuthMiddleware(), hConfigInfo.ConfigCleanup)
		configGroup.GET("/language_list", hConfigInfo.ConfigLanguageList)
	}
}
