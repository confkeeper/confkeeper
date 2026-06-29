package captcha

import (
	"confkeeper/utils"
	"confkeeper/utils/config"
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/mojocn/base64Captcha"
)

// Driver 验证码驱动 - 延迟初始化
var Driver *base64Captcha.DriverString
var Store base64Captcha.Store

// RedisStore 基于 Redis 的验证码存储实现
type RedisStore struct {
	expiration time.Duration
}

// redisCaptchaKey 生成 Redis 中存储验证码的 key
func redisCaptchaKey(id string) string {
	return fmt.Sprintf("captcha:%s", id)
}

// Set 将验证码存入 Redis
func (s *RedisStore) Set(id string, value string) error {
	if utils.RedisClient == nil {
		return fmt.Errorf("Redis 未连接")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return utils.RedisClient.Set(ctx, redisCaptchaKey(id), strings.ToLower(value), s.expiration).Err()
}

// Get 从 Redis 获取验证码，clear 为 true 时删除
func (s *RedisStore) Get(id string, clear bool) string {
	if utils.RedisClient == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := redisCaptchaKey(id)
	val, err := utils.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		utils.RedisClient.Del(ctx, key)
	}
	return val
}

// Verify 验证验证码
func (s *RedisStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v == strings.ToLower(answer)
}

// Init 初始化验证码配置
func Init() {
	Driver = base64Captcha.NewDriverString(
		60,                                  // 高度
		240,                                 // 宽度
		config.Cfg.Captcha.NoiseCount,       // 干扰数量
		config.Cfg.Captcha.InterferenceLine, // 同时显示直线和曲线干扰
		config.Cfg.Captcha.Length,           // 验证码长度
		config.Cfg.Captcha.CharacterSet,     // 字符集
		&color.RGBA{R: 240, G: 240, B: 240, A: 255}, // 背景颜色
		nil, // 字体存储
		[]string{"actionj.ttf", "wqy-microhei.ttc"}, // 字体列表
	)
	// 默认使用内存存储
	Store = base64Captcha.NewMemoryStore(10240, time.Duration(config.Cfg.Server.CaptchaExpireTime)*time.Minute)
	if !utils.EnableRedis || utils.RedisClient == nil {
		return
	}

	// 测试 Redis 连接是否可用
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := utils.RedisClient.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis 连接不可用，验证码继续使用内存存储")
		return
	}

	Store = &RedisStore{
		expiration: time.Duration(config.Cfg.Server.CaptchaExpireTime) * time.Minute,
	}
	slog.Info("验证码存储已切换到 Redis")
}

// ensure RedisStore implements base64Captcha.Store at compile time
var _ base64Captcha.Store = (*RedisStore)(nil)
