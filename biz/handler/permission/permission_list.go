package permission

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/response"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ListReq struct {
	Page     int32  `form:"page" binding:"required,min=1,max=1000"`
	PageSize int32  `form:"page_size" binding:"required,min=1,max=100"`
	Role     string `form:"role"`
}

type ListData struct {
	Role     string `json:"role"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

type ListResp struct {
	Code  response.Code `json:"code"`
	Msg   string        `json:"msg"`
	Total int64         `json:"total"`
	Data  []*ListData   `json:"data"`
}

// PermissionList 获取权限列表
//
//	@Tags			权限管理
//	@Summary		权限列表
//	@Description	获取角色的权限列表
//	@Accept			application/json
//	@Produce		application/json
//	@Param			role		query		string	true	"角色名"
//	@Param			page		query		int		false	"页码"	default(1)
//	@Param			page_size	query		int		false	"每页数量"	default(10)
//	@Success		200			{object}	ListResp
//	@Security		ApiKeyAuth
//	@router			/api/permission/list [GET]
func PermissionList(c *gin.Context) {
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

	offset := (req.Page - 1) * req.PageSize

	permissions, total, err := dal.GetRolePermissionsList(req.Role, int(offset), int(req.PageSize))
	if err != nil {
		c.JSON(http.StatusOK, &ListResp{
			Code: response.Code_DBErr,
			Msg:  "获取权限列表失败: " + err.Error(),
		})
		return
	}

	var permissionList []*ListData
	for _, p := range permissions {
		permissionList = append(permissionList, &ListData{
			Role:     p.Role,
			Resource: p.Resource,
			Action:   p.Action,
		})
	}

	resp.Code = response.Code_Success
	resp.Msg = "获取成功"
	resp.Total = total
	resp.Data = permissionList

	c.JSON(http.StatusOK, resp)
}
