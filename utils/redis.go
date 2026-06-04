package utils

import (
	"confkeeper/utils/config"
	"context"
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
func InitRedis() error {
	if !config.Cfg.Jwt.EnableMemory || config.Cfg.Jwt.Redis.Host == "" {
		EnableRedis = false
		slog.Info("Redis 未启用，JWT token 将存储在内存中")
		return nil
	}

	addr := fmt.Sprintf("%s:%d", config.Cfg.Jwt.Redis.Host, config.Cfg.Jwt.Redis.Port)
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Cfg.Jwt.Redis.Password,
		DB:       config.Cfg.Jwt.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	EnableRedis = true
	slog.Info("Redis 连接成功，JWT token 将存储在 Redis 中")
	return nil
}

// redisTokenKey 生成 Redis 中存储用户 token 列表的 key
func redisTokenKey(userid int) string {
	return fmt.Sprintf("jwt:tokens:%d", userid)
}

// RedisAddToken 将用户的 token 添加到 Redis 列表中，并限制最大会话数
func RedisAddToken(userid int, token string, maxSessions int) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis 未连接")
	}

	if maxSessions <= 0 {
		maxSessions = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := redisTokenKey(userid)

	pipe := RedisClient.TxPipeline()
	pipe.RPush(ctx, key, token)
	pipe.LTrim(ctx, key, int64(-maxSessions), -1)

	expireHours := time.Duration(config.Cfg.Jwt.ExpireTime*2) * time.Hour
	if expireHours < time.Hour {
		expireHours = time.Hour
	}
	pipe.Expire(ctx, key, expireHours)

	_, err := pipe.Exec(ctx)
	return err
}

// RedisLoadTokens 从 Redis 加载用户的 token 列表
func RedisLoadTokens(userid int) ([]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis 未连接")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return RedisClient.LRange(ctx, redisTokenKey(userid), 0, -1).Result()
}

// RedisRemoveToken 从 Redis 列表中删除指定的 token
func RedisRemoveToken(userid int, token string) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis 未连接")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return RedisClient.LRem(ctx, redisTokenKey(userid), 0, token).Err()
}
