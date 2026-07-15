package handler

import (
	"confkeeper/biz/dal"
	"confkeeper/internal/version"
	"confkeeper/utils/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping 测试网络接口
//
//	@Tags			测试
//	@Summary		测试网络接口
//	@Description	测试网络接口
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/api/ping [get]
func Ping(c *gin.Context) {
	err := dal.CheckDb()
	if err != nil {
		JSON(c, http.StatusOK, Code_DBErr, "数据库连接失败")
		return
	}
	JSON(c, http.StatusOK, Code_Success, "pong")
}

// ServerInfo 服务信息
//
//	@Tags			测试
//	@Summary		服务信息
//	@Description	服务信息
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/api/server_info [get]
func ServerInfo(c *gin.Context) {
	JSONData(c, http.StatusOK, Code_Success, "获取成功",
		map[string]any{
			"name":    config.Cfg.Server.Name,
			"version": version.Version,
			"ldap":    config.Cfg.Ldap.Enabled,
		})
}

// GetDemo 获取demo状态
//
//	@Tags			测试
//	@Summary		获取demo
//	@Description	获取demo
//	@Accept			application/json
//	@Produce		application/json
//	@Router			/api/is_demo [get]
func GetDemo(c *gin.Context) {
	JSONData(c, http.StatusOK, Code_Success, "获取成功",
		map[string]any{
			"is_demo": config.Cfg.Server.IsDemo,
		})
}
