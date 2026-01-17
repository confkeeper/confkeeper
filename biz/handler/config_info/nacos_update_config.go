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

type NacosUpdateReq struct {
	Tenant  string `form:"tenant" binding:"required,min=1,max=100"`
	DataId  string `form:"dataId" binding:"required,min=1,max=100"`
	Group   string `form:"group" binding:"required,min=1,max=100"`
	Type    string `form:"type" binding:"required,min=1,max=100"`
	Content string `form:"content" binding:"required"`
}

type NacosUpdateTokenReq struct {
	AccessToken string `form:"accessToken" binding:"required,min=1"`
}

// NacosUpdateConfig 更新/创建配置(nacos兼容)
//
//	@Tags			配置
//	@Tags			nacos兼容
//	@Summary		更新/创建配置(nacos兼容)
//	@Description	更新/创建配置(nacos兼容)
//	@Accept			application/x-www-form-urlencoded
//	@Produce		application/json
//	@Param			accessToken	query		string	true	"token"
//	@Param			tenant		formData	string	true	"tenant"
//	@Param			dataId		formData	string	true	"dataId"
//	@Param			group		formData	string	true	"group"
//	@Param			type		formData	string	true	"type"
//	@Param			content		formData	string	true	"content"
//	@Success		200			{object}	response.CommonResp
//	@router			/nacos/v1/cs/configs [POST]
func NacosUpdateConfig(c *gin.Context) {
	req := new(NacosUpdateReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	uriReq := new(NacosUpdateTokenReq)
	if err := c.ShouldBindQuery(uriReq); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	err := utils.ValidateShortTermToken(c, uriReq.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, &response.CommonResp{
			Code: response.Code_Err,
			Msg:  "token无效",
		})
		return
	}

	// 权限检查：管理员或有命名空间rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的rw权限
		hasPermission, err := mw.CheckNamespaceWritePermissionHTTP(c, req.Tenant)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_Unauthorized,
				Msg:  "没有发布配置的权限",
			})
			return
		}
	}

	// 检查命名空间是否存在
	exist, err := dal.IsTenantIdExists(req.Tenant)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if !exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Err,
			Msg:  "命名空间不存在",
		})
		return
	}

	// 判断该 tenant 下是否已存在该 dataId+group 的配置
	exists, err := dal.IsConfigInfoExists(req.DataId, req.Group, req.Tenant)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}

	var versionToCreate int
	if !exists {
		// 不存在则新增，version=1
		versionToCreate = 1
	} else {
		// 已存在则在该 tenant 作用域下取最大版本+1
		maxVersion, err := dal.GetMaxVersionByDataIdGroupAndTenant(req.DataId, req.Group, req.Tenant)
		if err != nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_DBErr,
				Msg:  "数据库查询错误: " + err.Error(),
			})
			return
		}
		versionToCreate = maxVersion + 1
	}

	cfg := &model.ConfigInfo{
		DataID:   req.DataId,
		GroupID:  req.Group,
		Content:  req.Content,
		TenantID: req.Tenant,
		Type:     req.Type,
		Version:  versionToCreate,
		Author:   c.GetString("username"),
	}

	if err = dal.CreateConfigInfo([]*model.ConfigInfo{cfg}); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "创建配置失败: " + err.Error(),
		})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "上传成功"

	c.JSON(http.StatusOK, resp)
	handler.IncConfigChange()
}
