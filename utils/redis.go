package utils

import (
	"confkeeper/utils/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gookit/slog"
)

var (
	// RedisClient 全局 Redis 客户端，仅在启用时初始化
	RedisClient *redis.Client
	// EnableRedis 是否启用 Redis 存储 JWT token
	EnableRedis bool
)

// InitRedis 初始化 Redis 连接，当 enable_memory 为 true 且配置了 Redis addr 时启用
func InitRedis() {
	if !config.Cfg.Jwt.EnableMemory || config.Cfg.Jwt.Redis.Addr == "" {
		EnableRedis = false
		slog.Info("Redis 未启用，JWT token 将存储在内存中")
		return
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Cfg.Jwt.Redis.Addr,
		Password: config.Cfg.Jwt.Redis.Password,
		DB:       config.Cfg.Jwt.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		EnableRedis = false
		slog.Errorf("Redis 连接失败: %v，JWT token 将存储在内存中", err)
		return
	}

	EnableRedis = true
	slog.Info("Redis 连接成功，JWT token 将存储在 Redis 中")
}

// redisTokenKey 生成 Redis 中存储用户 token 列表的 key
func redisTokenKey(userid int) string {
	return fmt.Sprintf("jwt:tokens:%d", userid)
}

// RedisStoreTokens 将用户的 token 列表存储到 Redis
func RedisStoreTokens(userid int, tokens []string) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis 未连接")
	}

	ctx := context.Background()
	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("序列化 token 列表失败: %w", err)
	}

	// 设置过期时间为 JWT 过期时间的 2 倍，确保 token 过期后 Redis 数据也能清理
	expireHours := time.Duration(config.Cfg.Jwt.ExpireTime*2) * time.Hour
	if expireHours < time.Hour {
		expireHours = time.Hour
	}

	return RedisClient.Set(ctx, redisTokenKey(userid), data, expireHours).Err()
}

// RedisLoadTokens 从 Redis 加载用户的 token 列表
func RedisLoadTokens(userid int) ([]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis 未连接")
	}

	ctx := context.Background()
	data, err := RedisClient.Get(ctx, redisTokenKey(userid)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // key 不存在，返回 nil
		}
		return nil, fmt.Errorf("从 Redis 获取 token 列表失败: %w", err)
	}

	var tokens []string
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("反序列化 token 列表失败: %w", err)
	}

	return tokens, nil
}

// RedisDeleteTokens 从 Redis 删除用户的 token 列表
func RedisDeleteTokens(userid int) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis 未连接")
	}

	ctx := context.Background()
	return RedisClient.Del(ctx, redisTokenKey(userid)).Err()
}
