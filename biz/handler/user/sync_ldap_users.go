package user

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/utils"
	"confkeeper/utils/config"
	"confkeeper/utils/ldap_client"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

// SyncLDAPUsersData 同步LDAP用户响应数据
type SyncLDAPUsersData struct {
	SyncedUsers    int `json:"synced_users"`     // 同步的用户数量
	TotalLDAPUsers int `json:"total_ldap_users"` // LDAP中的总用户数量
}

// SyncLDAPUsersResp 同步LDAP用户响应
type SyncLDAPUsersResp struct {
	Code handler.Code       `json:"code"` // 响应码
	Msg  string             `json:"msg"`  // 响应消息
	Data *SyncLDAPUsersData `json:"data"` // 响应数据
}

// SyncLDAPUsers 同步LDAP用户到本地数据库
//
//	@Tags			用户
//	@Summary		同步LDAP用户
//	@Description	从LDAP同步所有用户到本地数据库，仅添加本地不存在的用户
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	SyncLDAPUsersResp
//	@router			/api/user/sync-ldap [POST]
func SyncLDAPUsers(c *gin.Context) {
	if !config.Cfg.Ldap.Enabled {
		c.JSON(http.StatusOK, &SyncLDAPUsersResp{
			Code: handler.Code_Err,
			Msg:  "LDAP未启用",
		})
		return
	}

	// 检查管理员权限
	err := utils.IsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, &SyncLDAPUsersResp{
			Code: handler.Code_Unauthorized,
			Msg:  err.Error(),
		})
		return
	}
	// 从LDAP获取所有用户
	ldapUsers, err := ldap_client.GetAllLDAPUsers()
	if err != nil {
		slog.Errorf("从LDAP获取用户失败: %v", err)
		c.JSON(http.StatusOK, &SyncLDAPUsersResp{
			Code: handler.Code_Err,
			Msg:  "从LDAP获取用户失败",
		})
		return
	}

	// 先一次性查询数据库中所有用户，存储到map中以提高性能
	existingUsers, _, err := dal.GetAllUser()
	if err != nil {
		slog.Errorf("查询本地用户失败: %v", err)
		c.JSON(http.StatusOK, &SyncLDAPUsersResp{
			Code: handler.Code_DBErr,
			Msg:  "查询本地用户失败",
		})
		return
	}

	// 构建用户名到用户的映射
	existingUserMap := make(map[string]bool)
	for _, user := range existingUsers {
		existingUserMap[user.Username] = true
	}

	// 准备要添加的用户
	var usersToAdd []*model.User
	for _, ldapUser := range ldapUsers {
		username, ok := ldapUser["uid"]
		if !ok {
			continue
		}

		// 检查用户是否已存在
		if !existingUserMap[username] {
			// 用户不存在，添加到本地数据库
			// 密码字段为空，后续登录时会通过LDAP验证
			user := &model.User{
				Username: username,
				Password: "", // 密码为空
				Enable:   true,
			}
			usersToAdd = append(usersToAdd, user)
		}
	}

	// 批量添加用户
	if len(usersToAdd) > 0 {
		err = dal.CreateUser(usersToAdd)
		if err != nil {
			slog.Errorf("添加用户失败: %v", err)
			c.JSON(http.StatusOK, &SyncLDAPUsersResp{
				Code: handler.Code_DBErr,
				Msg:  "添加用户失败",
			})
			return
		}
	}

	c.JSON(http.StatusOK, &SyncLDAPUsersResp{
		Code: handler.Code_Success,
		Msg:  "同步完成",
		Data: &SyncLDAPUsersData{
			SyncedUsers:    len(usersToAdd),
			TotalLDAPUsers: len(ldapUsers),
		},
	})
}
