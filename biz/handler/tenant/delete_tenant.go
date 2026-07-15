package tenant

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteReq struct {
	ID string `uri:"id" binding:"required,min=1,max=255"`
}

type DeleteTenantResp struct {
	handler.CommonResp
}

// DeleteTenant 删除命名空间
//
//	@Tags			命名空间
//	@Summary		删除命名空间
//	@Description	删除命名空间
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string	true	"命名空间ID"
//	@Success		200	{object}	DeleteTenantResp
//	@Security		ApiKeyAuth
//	@router			/api/tenant/delete/{id} [DELETE]
func DeleteTenant(c *gin.Context) {
	req := new(DeleteReq)
	if err := c.ShouldBindUri(&req); err != nil {
		handler.ParamError(c, err)
		return
	}
	resp := new(DeleteTenantResp)

	// 检查租户下是否还有配置
	id, _ := strconv.Atoi(req.ID)
	tenantInfo, err := dal.GetTenantById(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, &DeleteTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "查询命名空间失败: " + err.Error(),
			},
		})
		return
	}
	if tenantInfo == nil {
		c.JSON(http.StatusOK, &DeleteTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Err,
				Msg:  "命名空间不存在",
			},
		})
		return
	}

	exist, err := dal.IsConfigInfoExistsByTenantId(tenantInfo.TenantID)
	if err != nil {
		c.JSON(http.StatusOK, &DeleteTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "检查配置失败: " + err.Error(),
			},
		})
		return
	}
	if exist {
		c.JSON(http.StatusOK, &DeleteTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Err,
				Msg:  "该命名空间下还有配置，不能删除",
			},
		})
		return
	}

	if err = dal.DeleteTenant(uint(id)); err != nil {
		c.JSON(http.StatusOK, &DeleteTenantResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "删除命名空间失败: " + err.Error(),
			},
		})
		return
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "删除命名空间成功",
	}

	c.JSON(http.StatusOK, resp)
}
