package config_info

import (
	"confkeeper/biz/handler"
	"confkeeper/utils/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LanguageListResp struct {
	handler.CommonResp
	Data []string `json:"data"`
}

// ConfigLanguageList 配置支持语言列表
//
//	@Tags			配置
//	@Summary		配置支持语言列表
//	@Description	配置支持语言列表
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	LanguageListResp
//	@Security		ApiKeyAuth
//	@router			/api/config/language_list [GET]
func ConfigLanguageList(c *gin.Context) {
	resp := new(LanguageListResp)

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "获取成功",
	}
	resp.Data = config.Cfg.Confkeeper.ConfigType

	c.JSON(http.StatusOK, resp)
}
