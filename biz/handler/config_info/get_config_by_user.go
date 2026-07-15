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
//	@Produce		application/json
//	@Param			username	query		string	true	"用户名"
//	@Param			password	query		string	true	"密码"
//	@Param			tenant		query		string	true	"命名空间"
//	@Param			dataId		query		string	true	"数据ID"
//	@Param			group		query		string	true	"分组ID"
//	@Success		200			{object}	handler.CommonJSONResp
//	@Failure		404			{object}	handler.CommonJSONResp	"配置不存在"
//	@Failure		500			{object}	handler.CommonJSONResp	"服务器错误"
//	@router			/api/config/get_by_user [GET]
//	@router			/api/config/get_by_user [HEAD]
func GetConfigByUser(c *gin.Context) {
	req := new(GetConfigByUserReq)
	if err := c.ShouldBindQuery(req); err != nil {
		handler.ParamError(c, err)
		return
	}

	userData, err := dal.UserLogin(req.Username)
	if err != nil {
		handler.JSON(c, http.StatusOK, handler.Code_DBErr, err.Error())
		return
	}
	if !utils.CheckPasswordHash(req.Password, userData.Password) {
		handler.JSON(c, http.StatusOK, handler.Code_PasswordErr, "密码错误")
		return
	}
	c.Set("userid", int(userData.ID))
	c.Set("username", userData.Username)

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
