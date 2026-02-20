package tenant

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	TenantId   string `json:"tenant_id" binding:"required,min=1,max=255"`
	TenantName string `json:"tenant_name" binding:"required,min=1,max=255"`
	TenantDesc string `json:"tenant_desc" binding:"required,min=1,max=255"`
}

type CreateTenantResp struct {
	handler.CommonResp
}

// CreateTenant 创建命名空间
//
//	@Tags			命名空间
//	@Summary		创建命名空间
//	@Description	创建命名空间
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		CreateReq	true	"创建命名空间请求参数"
//	@Success		200		{object}	CreateTenantResp
//	@Security		ApiKeyAuth
//	@router			/api/tenant/add [PUT]
func CreateTenant(c *gin.Context) {
	req := new(CreateReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(CreateTenantResp)

	// 检查命名空间是否已存在
	exist, err := dal.IsTenantIdExists(req.TenantId)
	if err != nil {
		c.JSON(http.StatusOK, &CreateTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "检查命名空间失败: " + err.Error(),
			},
		})
		return
	}
	if exist {
		c.JSON(http.StatusOK, &CreateTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_AlreadyExists,
				Msg:  "该命名空间已存在",
			},
		})
		return
	}

	t := &model.TenantInfo{
		TenantID:   req.TenantId,
		TenantName: req.TenantName,
		TenantDesc: req.TenantDesc,
	}

	if err = dal.CreateTenant([]*model.TenantInfo{t}); err != nil {
		c.JSON(http.StatusOK, &CreateTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "命名空间新建失败: " + err.Error(),
			},
		})
		return
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "新建命名空间成功",
	}

	c.JSON(http.StatusOK, resp)
}
