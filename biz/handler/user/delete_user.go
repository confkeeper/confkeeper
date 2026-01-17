package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteReq struct {
	UserId string `uri:"user_id" binding:"required,min=1,max=255"`
}

// DeleteUser 删除用户
//
//	@Tags			用户
//	@Summary		删除用户
//	@Description	删除用户
//	@Accept			application/json
//	@Produce		application/json
//	@Param			user_id	path		string	true	"用户ID"
//	@Success		200		{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/user/delete/{user_id} [DELETE]
func DeleteUser(c *gin.Context) {
	req := new(DeleteReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_Unauthorized,
			Msg:  err.Error(),
		})
		return
	}

	reqUserId, _ := strconv.Atoi(req.UserId)

	if reqUserId == 1 {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_Err, Msg: "不能删除管理员"})
		return
	}

	userId, _ := utils.GetUseridFromContext(c)
	if userId != 1 {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_Err, Msg: "非管理员账号没有权限"})
		return
	}

	if err = dal.DeleteUser(reqUserId); err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_DBErr, Msg: "删除用户失败: " + err.Error()})
		return
	}
	resp.Code = response.Code_Success
	resp.Msg = "用户" + req.UserId + "删除成功"

	c.JSON(http.StatusOK, resp)
}
