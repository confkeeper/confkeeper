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

type ListReq struct {
	Page     int32   `form:"page" binding:"required,min=1,max=1000"`
	PageSize int32   `form:"page_size" binding:"required,min=1,max=100"`
	DataId   *string `form:"data_id" binding:"omitempty,min=1,max=255"`
	GroupId  *string `form:"group_id" binding:"omitempty,min=1,max=255"`
	Type     *string `form:"type" binding:"omitempty,min=1,max=255"`
	TenantId string  `form:"tenant_id" binding:"required,min=1,max=100"`
}

type ListData struct {
	ConfigId   string `json:"config_id"`
	DataId     string `json:"data_id"`
	GroupId    string `json:"group_id"`
	Type       string `json:"type"`
	CreateTime string `json:"create_time"`
}

type ListResp struct {
	Code  response.Code `json:"code"`
	Msg   string        `json:"msg"`
	Total int64         `json:"total"`
	Data  []*ListData   `json:"data"`
}

// ConfigList 配置列表
//
//	@Tags			配置
//	@Summary		配置列表
//	@Description	配置列表
//	@Accept			application/json
//	@Produce		application/json
//	@Param			page		query		int		false	"页码"	default(1)
//	@Param			page_size	query		int		false	"每页数量"	default(10)
//	@Param			tenant_id	query		string	false	"命名空间id"
//	@Param			data_id		query		string	false	"配置id"
//	@Param			group_id	query		string	false	"组id"
//	@Param			type		query		string	false	"类型"
//	@Success		200			{object}	ListResp
//	@Security		ApiKeyAuth
//	@router			/api/config/list [GET]
func ConfigList(c *gin.Context) {
	req := new(ListReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(ListResp)

	// 权限检查：管理员或有命名空间r/rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		// 检查用户是否有命名空间的r或rw权限
		hasPermission, err := mw.CheckNamespaceReadOrWritePermissionHTTP(c, req.TenantId)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &ListResp{
				Code: response.Code_Unauthorized,
				Msg:  "没有查看配置列表的权限",
			})
			return
		}
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.Page - 1) * req.PageSize

	var DataId, GroupId, Type string
	if req.DataId != nil {
		DataId = *req.DataId
	}
	if req.GroupId != nil {
		GroupId = *req.GroupId
	}
	if req.Type != nil {
		Type = *req.Type
	}

	configInfos, total, err := dal.GetConfigInfoListWithMaxVersion(int(req.PageSize), int(offset), DataId, GroupId, Type, req.TenantId)
	if err != nil {
		c.JSON(http.StatusOK, &ListResp{
			Code: response.Code_DBErr,
			Msg:  "获取配置列表失败: " + err.Error(),
		})
		return
	}

	var configInfoList []*ListData
	for _, b := range configInfos {
		configInfoList = append(configInfoList, &ListData{
			ConfigId:   strconv.Itoa(int(b.ID)),
			DataId:     b.DataID,
			GroupId:    b.GroupID,
			Type:       b.Type,
			CreateTime: b.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	resp.Code = response.Code_Success
	resp.Msg = "获取成功"
	resp.Total = total
	resp.Data = configInfoList

	c.JSON(http.StatusOK, resp)
}
