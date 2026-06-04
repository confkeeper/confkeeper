package user

import (
	"confkeeper/biz/handler"
	"confkeeper/utils"
	"confkeeper/utils/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogoutResp struct {
	handler.CommonResp
}

// UserLogout 用户登出
//
//	@Tags			用户
//	@Summary		用户登出
//	@Description	用户登出，删除JWT token
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	LogoutResp
//	@router			/api/user/logout [POST]
func UserLogout(c *gin.Context) {
	// 如果启用了内存存储，则删除token
	if config.Cfg.Jwt.EnableMemory {
		// 从中间件获取 token 和 userid
		tokenString, _ := c.Get("token")
		userid, err := utils.GetUseridFromContext(c)
		if err != nil {
			c.JSON(http.StatusOK, &LogoutResp{
				CommonResp: handler.CommonResp{
					Code: handler.Code_Err,
					Msg:  "获取用户信息失败: " + err.Error(),
				},
			})
			return
		}

		if utils.EnableRedis {
			// 使用 Redis 删除 token
			_ = utils.RedisRemoveToken(userid, tokenString.(string))
		} else {
			// 从内存中删除token
			utils.TokenLock.Lock()
			defer utils.TokenLock.Unlock()

			if storedTokens, ok := utils.TokenStore.Load(userid); ok {
				if tokenList, ok := storedTokens.([]string); ok {
					// 从 token 列表中删除当前 token
					newTokenList := make([]string, 0)
					for _, t := range tokenList {
						if t != tokenString.(string) {
							newTokenList = append(newTokenList, t)
						}
					}

					// 如果还有其他 token，更新列表；否则删除整个条目
					if len(newTokenList) > 0 {
						utils.TokenStore.Store(userid, newTokenList)
					} else {
						utils.TokenStore.Delete(userid)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, &LogoutResp{
		CommonResp: handler.CommonResp{
			Code: handler.Code_Success,
			Msg:  "登出成功",
		},
	})
}
