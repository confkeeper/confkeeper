package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/mw"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ListVersionReq struct {
	ConfigId string `uri:"config_id" binding:"required,min=1,max=1000"`
}

type ListVersionData struct {
	ConfigId   string `json:"config_id"`
	TenantId   string `json:"tenant_id"`
	DataId     string `json:"data_id"`
	GroupId    string `json:"group_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Version    string `json:"version"`
	Author     string `json:"author"`
	CreateTime string `json:"create_time"`
}

type ListVersionResp struct {
	Code  response.Code      `json:"code"`
	Msg   string             `json:"msg"`
	Total int64              `json:"total"`
	Data  []*ListVersionData `json:"data"`
}

// ConfigVersion 获取配置的所有版本
//
//	@Tags			配置
//	@Summary		获取配置的所有版本
//	@Description	通过config_id查询data_id和group_id，然后查询所有不同版本，按版本字段倒序返回
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_id	path		string	true	"配置ID"
//	@Success		200			{object}	ListVersionResp
//	@Failure		400			{object}	ListVersionResp	"参数错误"
//	@Failure		404			{object}	ListVersionResp	"配置不存在"
//	@Failure		500			{object}	ListVersionResp	"服务器错误"
//	@router			/api/config/get_version/{config_id} [GET]
func ConfigVersion(c *gin.Context) {
	req := new(ListVersionReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(ListVersionResp)

	// 获取配置信息以检查权限
	configInfoData, err := dal.GetConfigInfoByID(req.ConfigId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ListVersionResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if configInfoData == nil {
		c.JSON(http.StatusNotFound, &ListVersionResp{
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
			c.JSON(http.StatusOK, &ListVersionResp{
				Code: response.Code_Unauthorized,
				Msg:  "没有查看配置版本的权限",
			})
			return
		}
	}

	// 根据data_id和group_id查询所有版本
	allVersions, err := dal.GetAllVersionsByDataIdAndGroup(configInfoData.DataID, configInfoData.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ListVersionResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}

	// 构建返回数据
	var versionList []*ListVersionData
	for _, version := range allVersions {
		versionList = append(versionList, &ListVersionData{
			ConfigId:   strconv.FormatUint(uint64(version.ID), 10),
			DataId:     version.DataID,
			GroupId:    version.GroupID,
			Version:    strconv.FormatUint(uint64(version.Version), 10),
			Content:    version.Content,
			Type:       version.Type,
			Author:     version.Author,
			CreateTime: version.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 创建新的响应结构体，使用repeated字段
	resp = &ListVersionResp{
		Code: response.Code_Success,
		Msg:  "获取配置版本成功",
		Data: versionList,
	}

	c.JSON(http.StatusOK, resp)
}
