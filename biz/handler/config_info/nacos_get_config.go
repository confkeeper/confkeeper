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
//	@Produce		application/json
//	@Param			accessToken	query		string	true	"token"
//	@Param			tenant		query		string	true	"tenant"
//	@Param			dataId		query		string	true	"dataId"
//	@Param			group		query		string	true	"group"
//	@Success		200			{object}	handler.CommonJSONResp
//	@Failure		404			{object}	handler.CommonJSONResp	"配置不存在"
//	@Failure		500			{object}	handler.CommonJSONResp	"服务器错误"
//	@router			/nacos/v1/cs/configs [GET]
//	@router			/nacos/v1/cs/configs [HEAD]
func NacosGetConfig(c *gin.Context) {
	req := new(NacosListReq)
	if err := c.ShouldBindQuery(req); err != nil {
		handler.ParamError(c, err)
		return
	}

	err := utils.ValidateShortTermToken(c, req.AccessToken)
	if err != nil {
		handler.JSON(c, http.StatusUnauthorized, handler.Code_Unauthorized, "token无效")
		return
	}

	// 检查用户是否启用
	if mw.CheckUserEnabled(c) {
		return
	}

	// 权限检查：管理员或有命名空间r/rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的r或rw权限
		hasPermission, err := mw.CheckNamespaceReadOrWritePermissionHTTP(c, req.Tenant)
		if err != nil || !hasPermission {
			handler.JSON(c, http.StatusUnauthorized, handler.Code_Unauthorized, "没有查看配置的权限")
			return
		}
	}

	// 检查命名空间是否存在
	exist, err := dal.IsTenantIdExists(req.Tenant)
	if err != nil {
		handler.JSON(c, http.StatusInternalServerError, handler.Code_DBErr, "数据库查询错误")
		return
	}
	if !exist {
		handler.JSON(c, http.StatusNotFound, handler.Code_Err, "命名空间不存在")
		return
	}

	// 获取最大版本的配置信息
	configInfoData, err := dal.GetConfigInfoByDataIdAndGroupWithMaxVersion(req.DataId, req.Group, req.Tenant)
	if err != nil {
		handler.JSON(c, http.StatusInternalServerError, handler.Code_DBErr, "数据库查询错误")
		return
	}
	if configInfoData == nil {
		handler.JSON(c, http.StatusNotFound, handler.Code_Err, "配置不存在")
		return
	}

	resp := configInfoData.Content

	handler.JSONData(c, http.StatusOK, handler.Code_Success, "获取成功", map[string]string{
		"content": resp,
	})
	handler.IncConfigRead()
}
