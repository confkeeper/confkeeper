package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/biz/mw"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	DataId   string `json:"data_id" binding:"required,min=1,max=255"`
	GroupId  string `json:"group_id" binding:"required,min=1,max=255"`
	TenantId string `json:"tenant_id" binding:"required,min=1,max=255"`
}

// CreateConfig 创建配置
//
//	@Tags			配置
//	@Summary		创建配置
//	@Description	创建配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		CreateReq	true	"配置信息"
//	@Success		200	{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/config/add [PUT]
func CreateConfig(c *gin.Context) {
	req := new(CreateReq)
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
				Msg:  "没有创建配置的权限",
			})
			return
		}
	}

	var exist bool

	// 检查命名空间是否不存在
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

	// 检查配置是否已存在
	exist, err = dal.IsConfigInfoExists(req.DataId, req.GroupId, req.TenantId)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查命名空间失败: " + err.Error(),
		})
		return
	}
	if exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_AlreadyExists,
			Msg:  "该配置已存在",
		})
		return
	}

	cfg := &model.ConfigInfo{
		DataID:   req.DataId,
		GroupID:  req.GroupId,
		Content:  "",
		TenantID: req.TenantId,
		Type:     "text",
		Version:  1, // 新配置版本为1
		Author:   c.GetString("username"),
	}

	if err = dal.CreateConfigInfo([]*model.ConfigInfo{cfg}); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_DBErr, Msg: "配置文件新建失败: " + err.Error()})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "新建配置文件成功"

	c.JSON(http.StatusOK, resp)
	handler.IncConfigChange()
}
