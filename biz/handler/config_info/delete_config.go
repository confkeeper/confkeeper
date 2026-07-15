package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/mw"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeleteReq struct {
	ConfigId string `uri:"config_id" binding:"required,min=1,max=100"`
}

type DeleteConfigResp struct {
	handler.CommonResp
}

// DeleteConfig 删除配置
//
//	@Tags			配置
//	@Summary		删除配置
//	@Description	删除配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_id	path		string	true	"配置ID"
//	@Success		200			{object}	DeleteConfigResp
//	@Security		ApiKeyAuth
//	@router			/api/config/delete/{config_id} [DELETE]
func DeleteConfig(c *gin.Context) {
	req := new(DeleteReq)
	if err := c.ShouldBindUri(req); err != nil {
		handler.ParamError(c, err)
		return
	}
	resp := new(DeleteConfigResp)

	// 获取配置信息以检查权限
	configInfoData, err := dal.GetConfigInfoByID(req.ConfigId)
	if err != nil {
		c.JSON(http.StatusOK, &DeleteConfigResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "查询配置信息失败: " + err.Error(),
			},
		})
		return
	}
	if configInfoData == nil {
		c.JSON(http.StatusOK, &DeleteConfigResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Err,
				Msg:  "配置不存在",
			},
		})
		return
	}

	// 权限检查：管理员或有命名空间rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的rw权限
		hasPermission, err := mw.CheckNamespaceWritePermissionHTTP(c, configInfoData.TenantID)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &DeleteConfigResp{
				CommonResp: handler.CommonResp{
					Code: handler.Code_Unauthorized,
					Msg:  "没有删除配置的权限",
				},
			})
			return
		}
	}

	if err = dal.DeleteConfigInfo(configInfoData.TenantID, configInfoData.DataID, configInfoData.GroupID); err != nil {
		c.JSON(http.StatusOK, &DeleteConfigResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "删除配置失败: " + err.Error(),
			},
		})
		return
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "配置删除成功",
	}

	c.JSON(http.StatusOK, resp)
}
