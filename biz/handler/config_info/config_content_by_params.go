package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/mw"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContentByParamsReq struct {
	TenantId string `form:"tenant_id" binding:"required,min=1,max=1000"`
	DataId   string `form:"data_id" binding:"required,min=1,max=100"`
	GroupId  string `form:"group_id" binding:"required,min=1,max=100"`
}

type ContentByParamsData struct {
	ConfigId   string `json:"config_id"`
	TenantId   string `json:"tenant_id"`
	DataId     string `json:"data_id"`
	GroupId    string `json:"group_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	CreateTime string `json:"create_time"`
}

type ContentByParamsResp struct {
	Code response.Code        `json:"code"`
	Msg  string               `json:"msg"`
	Data *ContentByParamsData `json:"data"`
}

// ConfigContentByParams 通过参数获取配置内容
//
//	@Tags			配置
//	@Summary		通过参数获取配置内容
//	@Description	通过tenant、dataId、group参数获取配置内容
//	@Accept			application/json
//	@Produce		application/json
//	@Param			tenant_id	query		string	true	"租户ID"
//	@Param			data_id		query		string	true	"数据ID"
//	@Param			group_id	query		string	true	"分组ID"
//	@Success		200			{object}	ContentByParamsResp
//	@Failure		400			{object}	ContentByParamsResp	"参数错误"
//	@Failure		404			{object}	ContentByParamsResp	"配置不存在"
//	@Failure		500			{object}	ContentByParamsResp	"服务器错误"
//	@router			/api/config/get [GET]
func ConfigContentByParams(c *gin.Context) {
	req := new(ContentByParamsReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(ContentByParamsResp)

	// 权限检查：管理员或有命名空间r/rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的r或rw权限
		hasPermission, err := mw.CheckNamespaceReadOrWritePermissionHTTP(c, req.TenantId)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &ContentByParamsResp{
				Code: response.Code_Unauthorized,
				Msg:  "没有查看配置的权限",
			})
			return
		}
	}

	// 检查命名空间是否存在
	exist, err := dal.IsTenantIdExists(req.TenantId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ContentByParamsResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if !exist {
		c.JSON(http.StatusNotFound, &ContentByParamsResp{
			Code: response.Code_Err,
			Msg:  "命名空间不存在",
		})
		return
	}

	// 获取最大版本的配置信息
	configInfoData, err := dal.GetConfigInfoByDataIdAndGroupWithMaxVersion(req.DataId, req.GroupId, req.TenantId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &ContentByParamsResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if configInfoData == nil {
		c.JSON(http.StatusNotFound, &ContentByParamsResp{
			Code: response.Code_Err,
			Msg:  "配置不存在",
		})
		return
	}

	// 返回配置详情
	resp.Code = response.Code_Success
	resp.Msg = "获取配置成功"
	resp.Data = &ContentByParamsData{
		ConfigId:   strconv.FormatUint(uint64(configInfoData.ID), 10),
		TenantId:   configInfoData.TenantID,
		DataId:     configInfoData.DataID,
		GroupId:    configInfoData.GroupID,
		Type:       configInfoData.Type,
		Content:    configInfoData.Content,
		CreateTime: configInfoData.CreateTime.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, resp)
	handler.IncConfigRead()
}
