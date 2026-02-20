package role

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateReq struct {
	Role      string   `json:"role" binding:"required,min=1,max=255"`
	Usernames []string `json:"usernames" binding:"required,min=1"`
}

type CreateRoleResp struct {
	handler.CommonResp
}

// CreateRole 创建角色
//
//	@Tags			角色管理
//	@Summary		创建角色
//	@Description	创建新的角色
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		CreateReq	true	"角色信息"
//	@Success		200	{object}	CreateRoleResp
//	@Security		ApiKeyAuth
//	@router			/api/role/add [PUT]
func CreateRole(c *gin.Context) {
	req := new(CreateReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(CreateRoleResp)

	// 检查是否为管理员
	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &CreateRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Unauthorized,
				Msg:  err.Error(),
			},
		})
		return
	}

	// 检查用户名是否存在
	missingUsernames, err := dal.FindMissingUsernames(req.Usernames)
	if err != nil {
		c.JSON(http.StatusOK, &CreateRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "检查用户名失败: " + err.Error(),
			},
		})
		return
	}
	if len(missingUsernames) > 0 {
		c.JSON(http.StatusOK, &CreateRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Err,
				Msg:  "用户 " + missingUsernames[0] + " 不存在",
			},
		})
		return
	}

	roles := make([]*model.Roles, len(req.Usernames))
	for i, username := range req.Usernames {
		roles[i] = &model.Roles{
			Username: username,
			Role:     req.Role,
		}
	}

	if err = dal.CreateRole(roles); err != nil {
		c.JSON(http.StatusOK, &CreateRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "角色创建失败: " + err.Error(),
			},
		})
		return
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "创建角色成功",
	}

	c.JSON(http.StatusOK, resp)
}
