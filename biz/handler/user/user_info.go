package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InfoReq struct {
	UserId string `uri:"user_id" binding:"required,min=1,max=1000"`
}

type InfoData struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Enable   bool   `json:"enable"`
}

type InfoResp struct {
	Code  response.Code `json:"code"`
	Msg   string        `json:"msg"`
	Total int64         `json:"total"`
	Data  *InfoData     `json:"data"`
}

// UserInfo 用户信息
//
//	@Tags			用户
//	@Summary		用户信息
//	@Description	用户信息
//	@Accept			application/json
//	@Produce		application/json
//	@Param			user_id	path		string	true	"用户ID"
//	@Success		200		{object}	InfoResp
//	@Security		ApiKeyAuth
//	@router			/api/user/info/{user_id} [GET]
func UserInfo(c *gin.Context) {
	req := new(InfoReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(InfoResp)

	userId, _ := strconv.Atoi(req.UserId)
	tokenUserId, _ := utils.GetUseridFromContext(c)

	if userId != tokenUserId {
		c.JSON(http.StatusOK, &InfoResp{Code: response.Code_Unauthorized, Msg: "不能修改获取别人"})
		return
	}

	// 获取用户信息
	userData, err := dal.GetUserByID(userId)
	if err != nil {
		c.JSON(http.StatusOK, &InfoResp{
			Code: response.Code_DBErr,
			Msg:  "数据库查询错误: " + err.Error(),
		})
		return
	}
	if userData == nil {
		c.JSON(http.StatusOK, &InfoResp{
			Code: response.Code_DBErr,
			Msg:  "用户未找到",
		})
		return
	}

	resp.Code = response.Code_Success
	resp.Msg = "用户信息更新成功"
	resp.Data = &InfoData{
		UserId:   strconv.Itoa(int(userData.ID)),
		Username: userData.Username,
		Enable:   userData.Enable,
	}

	c.JSON(http.StatusOK, resp)
}
