package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ListReq struct {
	Page     int32  `form:"page" binding:"required,min=1,max=1000"`
	PageSize int32  `form:"page_size" binding:"required,min=1,max=100"`
	Username string `form:"username" binding:"omitempty,min=1,max=255"`
}

type ListData struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Enable   bool   `json:"enable"`
}

type ListResp struct {
	Code  response.Code `json:"code"`
	Msg   string        `json:"msg"`
	Total int64         `json:"total"`
	Data  []*ListData   `json:"data"`
}

// UserList 用户列表
//
//	@Tags			用户
//	@Summary		用户列表
//	@Description	用户列表
//	@Accept			application/json
//	@Produce		application/json
//	@Param			page		query		int		false	"页码"	default(1)
//	@Param			page_size	query		int		false	"每页数量"	default(10)
//	@Param			username	query		string	false	"用户名"
//	@Param			email		query		string	false	"邮箱"
//	@Success		200			{object}	ListResp
//	@Security		ApiKeyAuth
//	@router			/api/user/list [GET]
func UserList(c *gin.Context) {
	req := new(ListReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(ListResp)

	// 检查管理员权限
	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &ListResp{
			Code: response.Code_Unauthorized,
			Msg:  err.Error(),
		})
		return
	}

	// 设置分页默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize

	// 获取用户列表和总数（转换分页参数类型）
	users, total, err := dal.GetUserList(int(req.PageSize), int(offset), req.Username)
	if err != nil {
		c.JSON(http.StatusOK, &ListResp{
			Code: response.Code_DBErr,
			Msg:  "获取用户列表失败: " + err.Error(),
		})
		return
	}

	// 转换响应格式
	var userList []*ListData
	for _, u := range users {
		userList = append(userList, &ListData{
			UserId:   strconv.Itoa(int(u.ID)),
			Username: u.Username,
			Enable:   u.Enable,
		})
	}

	resp.Code = response.Code_Success
	resp.Msg = "获取成功"
	resp.Total = total
	resp.Data = userList

	c.JSON(http.StatusOK, resp)
}
