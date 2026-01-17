package user

import (
	"confkeeper/biz/response"
	"confkeeper/utils/captcha"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// CaptchaResp 验证码响应
type CaptchaResp struct {
	Code response.Code `json:"code"`
	Msg  string        `json:"msg"`
	Data *CaptchaData  `json:"data"`
}

// CaptchaData 验证码数据
type CaptchaData struct {
	ID          string `json:"id"`
	Base64Image string `json:"base64_image"`
}

// GenerateCaptcha 生成验证码
//
//	@Tags			用户
//	@Summary		生成验证码
//	@Description	生成登录验证码
//	@Produce		application/json
//	@Success		200	{object}	CaptchaResp
//	@router			/api/user/captcha [GET]
func GenerateCaptcha(c *gin.Context) {
	// 创建验证码
	NewCaptcha := base64Captcha.NewCaptcha(captcha.Driver, captcha.Store)
	id, base64Image, _, err := NewCaptcha.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, &CaptchaResp{
			Code: response.Code_Err,
			Msg:  "生成验证码失败",
		})
		return
	}

	c.JSON(http.StatusOK, &CaptchaResp{
		Code: response.Code_Success,
		Msg:  "生成验证码成功",
		Data: &CaptchaData{
			ID:          id,
			Base64Image: base64Image,
		},
	})
}
