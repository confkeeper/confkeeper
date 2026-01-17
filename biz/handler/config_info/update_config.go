package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/biz/mw"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"confkeeper/utils/config"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

type UpdateReq struct {
	DataId  *string `json:"data_id" binding:"omitempty,min=1,max=255"`
	GroupId *string `json:"group_id" binding:"omitempty,min=1,max=255"`
	Content *string `json:"content" binding:"omitempty"`
	Type    *string `json:"type" binding:"omitempty,min=1,max=255"`
}

type UpdateUriReq struct {
	ConfigId string `uri:"config_id" binding:"required"`
}

// UpdateConfig 更新配置
//
//	@Tags			配置
//	@Summary		更新配置
//	@Description	更新配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_id	path		string		true	"配置ID"
//	@Param			req			body		UpdateReq	true	"配置信息"
//	@Success		200			{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/config/update/{config_id} [POST]
func UpdateConfig(c *gin.Context) {
	req := new(UpdateReq)
	uriReq := new(UpdateUriReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := c.ShouldBindUri(uriReq); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	// 获取配置信息以检查权限
	configInfoData, err := dal.GetConfigInfoByID(uriReq.ConfigId)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if configInfoData == nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Err,
			Msg:  "配置不存在",
		})
		return
	}

	// 权限检查：管理员或有命名空间rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的rw权限
		hasPermission, err := mw.CheckNamespaceWritePermissionHTTP(c, configInfoData.TenantID)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_Unauthorized,
				Msg:  "没有更新配置的权限",
			})
			return
		}
	}

	// 获取当前data_id、group_id、tenant_id的最大版本号
	maxVersion, err := dal.GetMaxVersionByDataIdGroupAndTenant(configInfoData.DataID, configInfoData.GroupID, configInfoData.TenantID)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}

	// 创建新版本配置
	newVersion := maxVersion + 1
	newConfig := &model.ConfigInfo{
		DataID:   configInfoData.DataID,
		GroupID:  configInfoData.GroupID,
		Content:  configInfoData.Content,
		TenantID: configInfoData.TenantID,
		Type:     configInfoData.Type,
		Version:  newVersion,
		Author:   c.GetString("username"),
	}

	// 使用请求中的新值
	if req.DataId != nil {
		newConfig.DataID = *req.DataId
	}
	if req.GroupId != nil {
		newConfig.GroupID = *req.GroupId
	}
	if req.Content != nil {
		newConfig.Content = *req.Content
	}
	if req.Type != nil {
		// 检查配置文件类型是否支持
		if !slices.Contains(config.Cfg.Confkeeper.ConfigType, *req.Type) {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_Unauthorized,
				Msg:  "配置文件类型不支持",
			})
			return
		}
		newConfig.Type = *req.Type
	}

	// 创建新配置记录
	err = dal.CreateConfigInfo([]*model.ConfigInfo{newConfig})
	if err != nil {
		c.JSON(http.StatusInternalServerError, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "创建配置版本失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	resp.Code = response.Code_Success
	resp.Msg = "配置信息更新成功"

	c.JSON(http.StatusOK, resp)
	handler.IncConfigChange()
}
