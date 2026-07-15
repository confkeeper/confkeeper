package role

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeleteReq struct {
	Role string `uri:"role" binding:"required,min=1,max=255"`
}

type DeleteRoleResp struct {
	handler.CommonResp
}

// DeleteRole 删除角色
//
//	@Tags			角色管理
//	@Summary		删除角色
//	@Description	删除指定角色及其所有权限
//	@Accept			application/json
//	@Produce		application/json
//	@Param			role	path		string	true	"角色名"
//	@Success		200		{object}	DeleteRoleResp
//	@Security		ApiKeyAuth
//	@router			/api/role/delete/{role} [DELETE]
func DeleteRole(c *gin.Context) {
	req := new(DeleteReq)
	if err := c.ShouldBindUri(req); err != nil {
		handler.ParamError(c, err)
		return
	}
	resp := new(DeleteRoleResp)

	// 删除角色及其所有权限
	if err := dal.DeleteRole(req.Role); err != nil {
		c.JSON(http.StatusOK, &DeleteRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "删除角色失败: " + err.Error(),
			},
		})
		return
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "删除角色成功",
	}

	c.JSON(http.StatusOK, resp)
}
