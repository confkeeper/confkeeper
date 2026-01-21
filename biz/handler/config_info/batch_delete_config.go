package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/mw"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BatchDeleteReq struct {
	ConfigIds []string `json:"config_ids" binding:"required,min=1,dive,min=1"`
}

// BatchDeleteConfig 批量删除配置
//
//	@Tags			配置
//	@Summary		批量删除配置
//	@Description	批量删除配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		BatchDeleteReq	true	"批量删除请求"
//	@Success		200	{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/config/batch_delete [DELETE]
func BatchDeleteConfig(c *gin.Context) {
	req := new(BatchDeleteReq)
	if err := c.ShouldBindJSON(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	// 遍历所有配置ID
	for _, configId := range req.ConfigIds {
		// 获取配置信息以检查权限
		configInfoData, err := dal.GetConfigInfoByID(configId)
		if err != nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_DBErr,
				Msg:  "查询配置信息失败: " + err.Error(),
			})
			return
		}
		if configInfoData == nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_Err,
				Msg:  "配置不存在: " + configId,
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
					Msg:  "没有删除配置的权限: " + configId,
				})
				return
			}
		}

		// 删除配置
		if err = dal.DeleteConfigInfo(configInfoData.TenantID, configInfoData.DataID, configInfoData.GroupID); err != nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_DBErr,
				Msg:  "删除配置失败: " + configId + " - " + err.Error(),
			})
			return
		}
	}

	resp.Code = response.Code_Success
	resp.Msg = "批量删除配置成功"

	c.JSON(http.StatusOK, resp)
}
