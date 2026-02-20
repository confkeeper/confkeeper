package mw

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckUserEnabled 检查用户是否启用，返回 true 表示已处理（需终止），false 表示通过
func CheckUserEnabled(c *gin.Context) bool {
	userID, ok := c.Get("userid")
	if !ok {
		return false
	}

	uid, ok := userID.(int)
	if !ok {
		return false
	}

	user, err := dal.GetUserByID(uid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code": handler.Code_Unauthorized,
			"msg":  "查询用户信息失败",
		})
		c.Abort()
		return true
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code": handler.Code_Unauthorized,
			"msg":  "用户不存在",
		})
		c.Abort()
		return true
	}
	if !user.Enable {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code": handler.Code_UserErr,
			"msg":  "用户已被禁用",
		})
		c.Abort()
		return true
	}
	return false
}

// JWTAuthMiddleware 鉴权中间件
func JWTAuthMiddleware(opts ...utils.AuthOptions) gin.HandlerFunc {
	// 初始化默认值（Go 结构体布尔字段默认就是 false）
	var opt utils.AuthOptions

	// 如果用户传了参数，就用用户传的第一个参数覆盖默认值
	if len(opts) > 0 {
		opt = opts[0]
	}
	return func(c *gin.Context) {
		// 获取 Authorization Header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": http.StatusUnauthorized,
				"msg":  "缺少token",
			})
			c.Abort() // 终止后续处理
			return
		}

		// 提取token（去除Bearer前缀）
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": http.StatusUnauthorized,
				"msg":  "token格式错误",
			})
			c.Abort() // 终止后续处理
			return
		}

		// 验证 token
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": http.StatusUnauthorized,
				"msg":  err.Error(),
			})
			c.Abort() // 终止后续处理
			return
		}

		// 检查短时token
		if opt.IsShortTerm {
			tokenType, ok := claims["token_type"].(string)
			if !ok || tokenType != "short_term" {
				c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code": http.StatusUnauthorized,
					"msg":  "没有权限",
				})
				c.Abort() // 终止后续处理
				return
			}
		}

		// 将 claims 保存到上下文
		for k, v := range claims {
			c.Set(k, v)
		}
		c.Set("userid", claims["userid"])
		c.Set("username", claims["username"])
		c.Set("token", token)

		if opt.CheckAdmin {
			// 检查是否为管理员
			err := utils.IsAdmin(c)
			if err != nil {
				c.JSON(http.StatusOK, map[string]interface{}{
					"code": http.StatusUnauthorized,
					"msg":  "不是管理员",
				})
				c.Abort() // 终止后续处理
				return
			}
		}

		// 检查用户是否启用
		if CheckUserEnabled(c) {
			return
		}

		// 如果验证通过，继续处理请求
		c.Next()
	}
}
