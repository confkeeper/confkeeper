package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/mw"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NacosListReq struct {
	AccessToken string `form:"accessToken" binding:"required,min=1"`
	Tenant      string `form:"tenant" binding:"required,min=1,max=100"`
	DataId      string `form:"dataId" binding:"required,min=1,max=100"`
	Group       string `form:"group" binding:"required,min=1,max=100"`
}

// NacosGetConfig 获取配置(nacos兼容)
//
//	@Tags			配置
//	@Tags			nacos兼容
//	@Summary		获取配置(nacos兼容)
//	@Description	获取配置(nacos兼容)
//	@Accept			application/json
//	@Produce		text/plain
//	@Param			accessToken	query		string	true	"token"
//	@Param			tenant		query		string	true	"tenant"
//	@Param			dataId		query		string	true	"dataId"
//	@Param			group		query		string	true	"group"
//	@Param			tenant		query		string	true	"tenant"
//	@Success		200			{string}	string	"配置内容"
//	@Failure		404			{string}	string	"配置不存在"
//	@Failure		500			{string}	string	"服务器错误"
//	@Success		200			{string}	string	""
//	@router			/nacos/v1/cs/configs [GET]
func NacosGetConfig(c *gin.Context) {
	req := new(NacosListReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	err := utils.ValidateShortTermToken(c, req.AccessToken)
	if err != nil {
		c.String(http.StatusUnauthorized, "token无效")
		return
	}

	// 权限检查：管理员或有命名空间r/rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的r或rw权限
		hasPermission, err := mw.CheckNamespaceReadOrWritePermissionHTTP(c, req.Tenant)
		if err != nil || !hasPermission {
			c.String(http.StatusUnauthorized, "没有查看配置的权限")
			return
		}
	}

	// 检查命名空间是否存在
	exist, err := dal.IsTenantIdExists(req.Tenant)
	if err != nil {
		c.String(http.StatusInternalServerError, "数据库查询错误")
		return
	}
	if !exist {
		c.String(http.StatusNotFound, "命名空间不存在")
		return
	}

	// 获取最大版本的配置信息
	configInfoData, err := dal.GetConfigInfoByDataIdAndGroupWithMaxVersion(req.DataId, req.Group, req.Tenant)
	if err != nil {
		c.String(http.StatusInternalServerError, "数据库查询错误")
		return
	}
	if configInfoData == nil {
		c.String(http.StatusNotFound, "配置不存在")
		return
	}

	resp := configInfoData.Content

	// 直接返回配置内容，符合nacos格式
	c.String(http.StatusOK, resp)
	handler.IncConfigRead()
}
