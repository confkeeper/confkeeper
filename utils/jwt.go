package utils

import (
	"confkeeper/utils/config"
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gookit/slog"
)

var (
	// JWT 配置
	jwtConfig = struct {
		Secret        []byte
		IdentityKey   string
		TokenHeadName string
		Issuer        string
	}{
		Secret:        []byte(config.Cfg.Jwt.Secret),
		IdentityKey:   "userid",
		TokenHeadName: "Bearer",
	}

	// TokenStore 用于内存存储 userid -> token
	TokenStore sync.Map
)

// GenerateToken 生成 JWT Token
func GenerateToken(userid uint, username string, expTime ...int) (string, error) {
	// 设置过期时间
	var expireTime time.Duration
	if len(expTime) > 0 {
		expireTime = time.Minute * time.Duration(expTime[0])
	} else {
		expireTime = time.Hour * time.Duration(config.Cfg.Jwt.ExpireTime)
	}

	// 创建 claims
	claims := jwt.MapClaims{
		"userid":     int(userid),
		"username":   username,
		"iss":        config.Cfg.Server.Name,
		"exp":        time.Now().Add(expireTime).Unix(),
		"orig_iat":   time.Now().Unix(),
		"token_type": "access", // 默认为访问令牌
	}

	// 如果是短期令牌，添加标识
	if len(expTime) > 0 && expTime[0] > 0 && expTime[0] < 5 { // 小于5分钟的认为是短期令牌
		claims["token_type"] = "short_term"
	}

	// 生成 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtConfig.Secret)
	if err != nil {
		return "", err
	}

	// Store token in memory if enabled
	if config.Cfg.Jwt.EnableMemory {
		TokenStore.Store(int(userid), tokenString)
	}

	return tokenString, nil
}

// ParseToken 解析并验证 JWT Token
func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	// 解析 token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtConfig.Secret, nil
	})

	if err != nil {
		errMsg := err.Error()

		// 检查各种可能的签名错误消息
		if strings.Contains(errMsg, "signature") && strings.Contains(errMsg, "invalid") {
			return nil, fmt.Errorf("令牌签名验证失败，请重新登录")
		}

		if strings.Contains(errMsg, "token is expired") {
			return nil, fmt.Errorf("令牌已过期，请重新登录")
		}

		if strings.Contains(errMsg, "token is not valid yet") {
			return nil, fmt.Errorf("令牌尚未生效")
		}

		if strings.Contains(errMsg, "token is malformed") {
			return nil, fmt.Errorf("令牌格式错误")
		}

		return nil, fmt.Errorf("身份验证失败: %v", err)
	}

	// 验证 token 是否有效
	if !token.Valid {
		return nil, fmt.Errorf("token 无效")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("无法解析 claims")
	}

	// 验证 issuer
	if iss, ok := claims["iss"].(string); !ok || iss != config.Cfg.Server.Name {
		return nil, fmt.Errorf("issuer 不匹配")
	}

	// 验证过期时间
	if exp, ok := claims["exp"].(float64); !ok || int64(exp) < time.Now().Unix() {
		return nil, fmt.Errorf("token 已过期")
	}

	// userid 转换为 int
	var userid int
	if useridVal, ok := claims["userid"]; ok {
		if uid, ok := useridVal.(float64); ok {
			userid = int(uid)
			claims["userid"] = userid
		} else if uid, ok := useridVal.(int); ok {
			userid = uid
		}
	}

	// 验证 username
	if _, ok := claims["username"].(string); !ok {
		return nil, fmt.Errorf("token 缺少 username")
	}

	// Verify against in-memory store if enabled
	if config.Cfg.Jwt.EnableMemory {
		if storedToken, ok := TokenStore.Load(userid); ok {
			if storedTokenStr, ok := storedToken.(string); ok {
				if storedTokenStr != tokenStr {
					// Token mismatch (e.g. user logged in elsewhere or token revoked)
					return nil, fmt.Errorf("令牌无效或已失效，请重新登录")
				}
			} else {
				return nil, fmt.Errorf("服务器内部错误: 令牌存储格式错误")
			}
		} else {
			// No token found for this user in memory (perhaps server restarted or never logged in)
			return nil, fmt.Errorf("令牌不存在或已失效，请重新登录")
		}
	}

	return claims, nil
}

// GetUsernameFromContext 从上下文中提取用户名
func GetUsernameFromContext(c *gin.Context) (string, error) {
	usernameVal, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("未找到用户名")
	}

	username, ok := usernameVal.(string)
	if !ok {
		return "", fmt.Errorf("用户名类型错误")
	}

	return username, nil
}

// GetUseridFromContext 从上下文中提取用户ID
func GetUseridFromContext(c *gin.Context) (int, error) {
	useridVal, exists := c.Get("userid")
	if !exists {
		return 0, fmt.Errorf("未找到用户ID")
	}

	userid, ok := useridVal.(int)
	if !ok {
		return 0, fmt.Errorf("userid 类型错误")
	}

	return userid, nil
}

// IsAdmin 判断是否为管理员
func IsAdmin(c *gin.Context) error {
	userid, err := GetUseridFromContext(c)
	if err != nil {
		return err
	}

	if userid != 1 {
		slog.Infof("当前用户ID: %d", userid)
		return fmt.Errorf("不是管理员，没有权限")
	}

	return nil
}

func ValidateShortTermToken(c *gin.Context, token string) error {
	// 验证 token
	claims, err := ParseToken(token)
	if err != nil {
		return fmt.Errorf("token 无效: %v", err)
	}
	// 检查短时token
	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "short_term" {
		return fmt.Errorf("没有权限")
	}

	// 保存 claims 到 gin.Context
	for k, v := range claims {
		c.Set(k, v)
	}
	c.Set("userid", claims["userid"])
	c.Set("username", claims["username"])

	return nil
}
