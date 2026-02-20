package role

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EditReq struct {
	Role      string   `json:"role" binding:"required,min=1,max=255"`
	Usernames []string `json:"usernames" binding:"required,min=1"`
}

type EditRoleResp struct {
	handler.CommonResp
}

// EditRole 编辑角色下的用户列表
//
//	@Tags			角色管理
//	@Summary		编辑角色
//	@Description	全量更新角色关联的用户列表
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		EditReq	true	"编辑请求信息"
//	@Success		200	{object}	EditRoleResp
//	@Security		ApiKeyAuth
//	@router			/api/role/edit [PUT]
func EditRole(c *gin.Context) {
	req := new(EditReq)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	resp := new(EditRoleResp)

	// 检查是否为管理员
	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &EditRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Unauthorized,
				Msg:  err.Error(),
			},
		})
		return
	}

	// 预查角色是否存在？
	roleExist, err := dal.IsRoleExistsInRoles(req.Role)
	if err != nil {
		c.JSON(http.StatusOK, &EditRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "检查角色失败: " + err.Error(),
			},
		})
		return
	}
	if !roleExist && len(req.Usernames) > 0 {
		// 为了防止新建不存在的角色，如果角色完全不存在但是传了username过来，我们需要限制不允许通过 edit 凭空造角色吗？
		// 实际上允许通过 edit 建立新角色也是可以的，但语义上最好保持和原逻辑一致，若角色不存在则报错。
		c.JSON(http.StatusOK, &EditRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Err,
				Msg:  "需要编辑的角色不存在",
			},
		})
		return
	}

	// 检查每一个用户名是否真实存在
	for _, username := range req.Usernames {
		exist, err := dal.IsUsernameExists(username)
		if err != nil {
			c.JSON(http.StatusOK, &EditRoleResp{
				CommonResp: handler.CommonResp{
					Code: handler.Code_DBErr,
					Msg:  "检查用户名失败: " + err.Error(),
				},
			})
			return
		}
		if !exist {
			c.JSON(http.StatusOK, &EditRoleResp{
				CommonResp: handler.CommonResp{
					Code: handler.Code_Err,
					Msg:  "用户 " + username + " 不存在",
				},
			})
			return
		}
	}

	if err = dal.UpdateRoleUsers(req.Role, req.Usernames); err != nil {
		c.JSON(http.StatusOK, &EditRoleResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "更新角色失败: " + err.Error(),
			},
		})
		return
	}

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "更新角色成功",
	}

	c.JSON(http.StatusOK, resp)
}
