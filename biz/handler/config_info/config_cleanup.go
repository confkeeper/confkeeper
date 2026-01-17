package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ConfigCleanup 清理配置旧版本
//
//	@Tags			配置
//	@Summary		清理配置旧版本
//	@Description	清理配置旧版本
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/config/cleanup [POST]
func ConfigCleanup(c *gin.Context) {
	resp := new(response.CommonResp)

	// 权限检查：仅管理员可执行清理操作
	if err := utils.IsAdmin(c); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Unauthorized,
			Msg:  "只有管理员可以执行清理操作",
		})
		return
	}

	// 执行清理操作
	if err := dal.ClearOldConfigVersions(); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "清理配置版本失败: " + err.Error(),
		})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "配置版本清理成功"

	c.JSON(http.StatusOK, resp)
}
