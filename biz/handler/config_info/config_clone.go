package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/biz/mw"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CloneItems struct {
	ConfigId string `json:"config_id" binding:"required,min=1,max=100"`
	DataId   string `json:"data_id" binding:"required,min=1,max=100"`
	GroupId  string `json:"group_id" binding:"required,min=1,max=100"`
}

type CloneReq struct {
	TenantId string        `json:"tenant_id"`
	Items    []*CloneItems `json:"items"`
}

// ConfigClone 克隆配置
//
//	@Tags			配置
//	@Summary		克隆配置
//	@Description	根据提供的配置项列表进行克隆操作。先用data_id、group_id和tenant_id查询，如果不存在则用config_id查询原配置并插入新配置，version设为1
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		CloneReq	true	"克隆配置请求参数"
//	@Success		200	{object}	response.CommonResp
//	@Failure		400	{object}	response.CommonResp	"参数错误"
//	@Failure		401	{object}	response.CommonResp	"无权限"
//	@Failure		404	{object}	response.CommonResp	"命名空间不存在"
//	@Failure		500	{object}	response.CommonResp	"服务器错误"
//	@Security		ApiKeyAuth
//	@router			/api/config/clone [POST]
func ConfigClone(c *gin.Context) {
	req := new(CloneReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	// 权限检查：管理员或有命名空间rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的rw权限
		hasPermission, err := mw.CheckNamespaceWritePermissionHTTP(c, req.TenantId)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_Unauthorized,
				Msg:  "没有克隆配置的权限",
			})
			return
		}
	}

	// 检查命名空间是否存在
	exist, err := dal.IsTenantIdExists(req.TenantId)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查命名空间失败: " + err.Error(),
		})
		return
	}
	if !exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_AlreadyExists,
			Msg:  "该命名空间不存在",
		})
		return
	}

	// 处理items
	var configsToCreate []*model.ConfigInfo

	for _, item := range req.Items {
		// 用config_id查询原配置
		originalConfig, err := dal.GetConfigInfoByID(item.ConfigId)
		if err != nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_DBErr,
				Msg:  "查询原配置失败: " + err.Error(),
			})
			return
		}
		if originalConfig == nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_Err,
				Msg:  fmt.Sprintf("配置不存在: config_id=%s", item.ConfigId),
			})
			return
		}

		// 创建新配置，version设为1
		newConfig := &model.ConfigInfo{
			DataID:   item.DataId,
			GroupID:  item.GroupId,
			Content:  originalConfig.Content,
			TenantID: req.TenantId,
			Type:     originalConfig.Type,
			Version:  1,
			Author:   c.GetString("username"),
		}
		configsToCreate = append(configsToCreate, newConfig)
	}

	// 批量创建新配置
	if len(configsToCreate) > 0 {
		if err = dal.CreateConfigInfo(configsToCreate); err != nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_DBErr,
				Msg:  "创建配置失败: " + err.Error(),
			})
			return
		}
	}

	resp.Code = response.Code_Success
	resp.Msg = "配置克隆成功"

	c.JSON(http.StatusOK, resp)
	handler.IncConfigChange()
}
