package role

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DeleteReq struct {
	Role string `uri:"role" binding:"required,min=1,max=255"`
}

// DeleteRole 删除角色
//
//	@Tags			角色管理
//	@Summary		删除角色
//	@Description	删除指定角色及其所有权限
//	@Accept			application/json
//	@Produce		application/json
//	@Param			role	path		string	true	"角色名"
//	@Success		200		{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/role/delete/{role} [DELETE]
func DeleteRole(c *gin.Context) {
	req := new(DeleteReq)
	if err := c.ShouldBindUri(req); err != nil {
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

	// 删除角色及其所有权限
	if err = dal.DeleteRole(req.Role); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_DBErr, Msg: "删除角色失败: " + err.Error()})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "删除角色成功"

	c.JSON(http.StatusOK, resp)
}
