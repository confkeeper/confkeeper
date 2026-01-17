package captcha

import (
	"confkeeper/utils/config"
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
)

// Driver 验证码驱动 - 延迟初始化
var Driver *base64Captcha.DriverString
var Store base64Captcha.Store

// Init 初始化验证码配置
func Init() {
	Driver = base64Captcha.NewDriverString(
		60,                            // 高度
		240,                           // 宽度
		config.Cfg.Captcha.NoiseCount, // 干扰数量
		2,                             // 同时显示直线和曲线干扰
		config.Cfg.Captcha.Length,     // 验证码长度
		"abdefghjkmnpqrstuvwxyABDEFHJKMNPQRSTUVWXY023456789", // 字符集
		&color.RGBA{R: 240, G: 240, B: 240, A: 255},          // 背景颜色
		nil, // 字体存储
		[]string{"actionj.ttf", "wqy-microhei.ttc"}, // 字体列表
	)
	// 验证码存储 - 设置5分钟过期
	Store = base64Captcha.NewMemoryStore(10240, time.Duration(config.Cfg.Server.CaptchaExpireTime)*time.Minute)
}
