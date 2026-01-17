package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/mw"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ContentReq struct {
	ConfigId string `uri:"config_id" binding:"required"`
}

type ContentData struct {
	ConfigId string `json:"config_id"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	DataId   string `json:"data_id"`
	GroupId  string `json:"group_id"`
	TenantId string `json:"tenant_id"`
}

type ContentResp struct {
	Code  response.Code `json:"code"`
	Msg   string        `json:"msg"`
	Total int64         `json:"total"`
	Data  *ContentData  `json:"data"`
}

// ConfigContent 获取配置详情
//
//	@Tags			配置
//	@Summary		获取配置详情
//	@Description	获取配置详情
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_id	path		string	true	"配置ID"
//	@Success		200			{object}	ContentResp
//	@Security		ApiKeyAuth
//	@router			/api/config/get/{config_id} [GET]
func ConfigContent(c *gin.Context) {
	req := new(ContentReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(ContentResp)

	// 获取配置信息以检查权限
	configInfoData, err := dal.GetConfigInfoByID(req.ConfigId)
	if err != nil {
		c.JSON(http.StatusOK, &ContentResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if configInfoData == nil {
		c.JSON(http.StatusOK, &ContentResp{
			Code: response.Code_Err,
			Msg:  "配置不存在",
		})
		return
	}

	// 权限检查：管理员或有命名空间r/rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的r或rw权限
		hasPermission, err := mw.CheckNamespaceReadOrWritePermissionHTTP(c, configInfoData.TenantID)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &ContentResp{
				Code: response.Code_Unauthorized,
				Msg:  "没有查看配置的权限",
			})
			return
		}
	}

	resp.Code = response.Code_Success
	resp.Msg = "获取配置详情成功"
	resp.Data = &ContentData{
		ConfigId: req.ConfigId,
		Content:  configInfoData.Content,
		Type:     configInfoData.Type,
		DataId:   configInfoData.DataID,
		GroupId:  configInfoData.GroupID,
		TenantId: configInfoData.TenantID,
	}

	c.JSON(http.StatusOK, resp)
	handler.IncConfigRead()
}
