package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UpdateReq struct {
	Username *string `json:"username" binding:"omitempty,min=1,max=255"`
	Enable   *bool   `json:"enable" binding:"omitempty"`
}

type UpdateUriReq struct {
	UserId string `uri:"user_id" binding:"required"`
}

// UpdateUser 更新用户
//
//	@Tags			用户
//	@Summary		更新用户
//	@Description	更新用户
//	@Accept			application/json
//	@Produce		application/json
//	@Param			user_id	path		string		true	"用户ID"
//	@Param			req		body		UpdateReq	true	"更新信息"
//	@Success		200		{object}	response.CommonResp
//	@Security		ApiKeyAuth
//	@router			/api/user/update/{user_id} [POST]
func UpdateUser(c *gin.Context) {
	req := new(UpdateReq)
	uriReq := new(UpdateUriReq)
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
			c.JSON(http.StatusOK, &response.CommonResp{Code: response.Code_Unauthorized, Msg: "不能修改别人的信息"})
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

	// 更新用户名等其他字段
	if req.Username != nil {
		userData.Username = *req.Username
		// 先检查用户名是否已存在
		exist, err := dal.IsUsernameExists(*req.Username)
		if err != nil {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_DBErr,
				Msg:  "检查用户名失败: " + err.Error(),
			})
			return
		}
		if exist && userData.Username != *req.Username {
			c.JSON(http.StatusOK, &response.CommonResp{
				Code: response.Code_AlreadyExists,
				Msg:  "该用户已存在",
			})
			return
		}
	}

	if req.Enable != nil {
		userData.Enable = *req.Enable
	}

	// 方法保存数据
	err = dal.UpdateUser(userData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &response.CommonResp{
			Code: response.Code_DBErr,
			Msg:  "更新用户信息失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	resp.Code = response.Code_Success
	resp.Msg = "用户信息更新成功"

	c.JSON(http.StatusOK, resp)
}
