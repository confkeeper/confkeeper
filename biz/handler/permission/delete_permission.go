package permission

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeleteReq struct {
	Role     string `form:"role" binding:"required,min=1,max=255"`
	Action   string `form:"action" binding:"required,min=1,max=255"`
	Resource string `form:"resource" binding:"required,min=1,max=255"`
}

// DeletePermission 删除权限
//
//	@Tags			权限管理
//	@Summary		删除权限
//	@Description	删除角色的指定权限
//	@Accept			application/json
//	@Produce		application/json
//	@Param			role		query		string	true	"角色名"
//	@Param			resource	query		string	true	"资源路径"
//	@Param			action		query		string	true	"操作类型"
//	@Success		200			{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/permission/delete [DELETE]
func DeletePermission(c *gin.Context) {
	req := new(DeleteReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	// 检查是否为管理员
	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Unauthorized,
			Msg:  err.Error(),
		})
		return
	}

	// 检查要删除的权限是否存在
	exist, err := dal.IsPermissionExists(req.Role, req.Resource, req.Action)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "检查权限失败: " + err.Error(),
		})
		return
	}
	if !exist {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Err,
			Msg:  "权限不存在",
		})
		return
	}

	// 删除权限
	if err = dal.RemoveRolePermission(req.Role, req.Resource, req.Action); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_DBErr, Msg: "删除权限失败: " + err.Error()})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "删除权限成功"

	c.JSON(http.StatusOK, resp)
}
