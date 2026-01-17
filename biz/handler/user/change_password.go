package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChangePasswdReq struct {
	Password string `json:"password" binding:"required,min=1,max=255"`
}

type ChangePasswdUriReq struct {
	UserId string `uri:"user_id" binding:"required"`
}

// ChangePasswd 修改用户密码
//
//	@Tags			用户
//	@Summary		修改用户密码
//	@Description	修改用户密码
//	@Accept			application/json
//	@Produce		application/json
//	@Param			user_id	path		string			true	"用户ID"
//	@Param			req		body		ChangePasswdReq	true	"密码信息"
//	@Success		200		{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/user/change_passwd/{user_id} [POST]
func ChangePasswd(c *gin.Context) {
	req := new(ChangePasswdReq)
	uriReq := new(ChangePasswdUriReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := c.ShouldBindUri(uriReq); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(response.CommonResp)

	userId, _ := strconv.Atoi(uriReq.UserId)
	tokenUserId, _ := utils.GetUseridFromContext(c)

	if userId != tokenUserId {
		if tokenUserId != 1 {
			c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_Unauthorized, Msg: "不能修改别人的密码"})
			return
		}
	}

	// 获取用户信息
	userData, err := dal.GetUserByID(userId)
	if err != nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if userData == nil {
		c.JSON(http.StatusOK, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "用户未找到",
		})
		return
	}

	userData.Password = utils.MD5(req.Password)

	// 方法保存数据
	err = dal.UpdateUser(userData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "修改密码失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	resp.Code = response.Code_Success
	resp.Msg = "密码更新成功"

	c.JSON(http.StatusOK, resp)
}
