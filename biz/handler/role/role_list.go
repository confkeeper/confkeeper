package role

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ListReq struct {
	Page     int32 `form:"page" binding:"required,min=1,max=1000"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=100"`
}

type ListData struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type ListResp struct {
	Code  response.Code `json:"code"`
	Msg   string        `json:"msg"`
	Total int64         `json:"total"`
	Data  []*ListData   `json:"data"`
}

// RoleList 获取角色列表
//
//	@Tags			角色管理
//	@Summary		角色列表
//	@Description	获取所有角色列表
//	@Accept			application/json
//	@Produce		application/json
//	@Param			page		query		int	false	"页码"	default(1)
//	@Param			page_size	query		int	false	"每页数量"	default(10)
//	@Success		200			{object}	ListResp
//	@Security		ApiKeyAuth
//	@router			/api/role/list [GET]
func RoleList(c *gin.Context) {
	req := new(ListReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(ListResp)

	// 检查是否为管理员
	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &ListResp{
			Code: response.Code_Unauthorized,
			Msg:  err.Error(),
		})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize

	roles, total, err := dal.GetAllRolesWithPagination(int(req.PageSize), int(offset))
	if err != nil {
		c.JSON(http.StatusOK, &ListResp{
			Code: response.Code_DBErr,
			Msg:  "获取角色列表失败: " + err.Error(),
		})
		return
	}

	var roleList []*ListData
	for _, r := range roles {
		roleList = append(roleList, &ListData{
			Role:     r.Role,
			Username: r.Username,
		})
	}

	resp.Code = response.Code_Success
	resp.Msg = "获取成功"
	resp.Total = total
	resp.Data = roleList

	c.JSON(http.StatusOK, resp)
}
