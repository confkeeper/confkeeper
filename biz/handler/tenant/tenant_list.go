package tenant

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ListReq struct {
	Page     int32 `form:"page" binding:"required,min=1,max=1000"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=100"`
}

type ListData struct {
	Id         string `json:"id"`
	TenantId   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	TenantDesc string `json:"tenant_desc"`
}

type ListResp struct {
	handler.CommonResp
	Total int64       `json:"total"`
	Data  []*ListData `json:"data"`
}

// TenantList 命名空间列表
//
//	@Tags			命名空间
//	@Summary		命名空间列表
//	@Description	命名空间列表
//	@Accept			application/json
//	@Produce		application/json
//	@Param			page		query		int	false	"页码"	default(1)
//	@Param			page_size	query		int	false	"每页数量"	default(10)
//	@Success		200			{object}	ListResp
//	@Security		ApiKeyAuth
//	@router			/api/tenant/list [GET]
func TenantList(c *gin.Context) {
	req := new(ListReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(ListResp)
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.Page - 1) * req.PageSize

	tenants, total, err := dal.GetTenantList(int(req.PageSize), int(offset))
	if err != nil {
		c.JSON(http.StatusOK, &ListResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "获取命名空间列表失败: " + err.Error(),
			},
		})
		return
	}

	var tenantList []*ListData
	for _, b := range tenants {
		tenantList = append(tenantList, &ListData{
			Id:         strconv.Itoa(int(b.ID)),
			TenantId:   b.TenantID,
			TenantName: b.TenantName,
			TenantDesc: b.TenantDesc,
		})
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "获取成功",
	}
	resp.Total = total
	resp.Data = tenantList

	c.JSON(http.StatusOK, resp)
}
