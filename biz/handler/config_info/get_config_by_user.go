package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/mw"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetConfigByUserReq struct {
	Username string `form:"username" binding:"required,min=1,max=255"`
	Password string `form:"password" binding:"required,min=1,max=255"`
	Tenant   string `form:"tenant" binding:"required,min=1,max=100"`
	DataId   string `form:"dataId" binding:"required,min=1,max=100"`
	Group    string `form:"group" binding:"required,min=1,max=100"`
}

// GetConfigByUser 直接使用账号获取配置
//
//	@Tags			配置
//	@Summary		直接使用账号获取配置
//	@Description	直接使用账号获取配置
//	@Accept			application/json
//	@Produce		text/plain
//	@Param			username	query		string	true	"用户名"
//	@Param			password	query		string	true	"密码"
//	@Param			tenant		query		string	true	"租户ID"
//	@Param			dataId		query		string	true	"数据ID"
//	@Param			group		query		string	true	"分组ID"
//	@Success		200			{string}	string	"配置内容"
//	@Failure		404			{string}	string	"配置不存在"
//	@Failure		500			{string}	string	"服务器错误"
//	@Success		200			{string}	string	""
//	@router			/api/config/get_by_user [GET]
func GetConfigByUser(c *gin.Context) {
	req := new(GetConfigByUserReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	userData, err := dal.UserLogin(req.Username)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	if utils.MD5(req.Password) != userData.Password {
		c.String(http.StatusOK, "密码错误")
		return
	}
	c.Set("userid", int(userData.ID))
	c.Set("username", userData.Username)

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
