package mw

import (
	"confkeeper/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 鉴权中间件
func JWTAuthMiddleware(isShortTerm ...bool) gin.HandlerFunc {
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
		if len(isShortTerm) > 0 && isShortTerm[0] {
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

		// 如果验证通过，继续处理请求
		c.Next()
	}
}
