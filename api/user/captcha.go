package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"log"
	"rest/response"
)

// GenerateCaptcha 生成验证码并存入 Redis（不限制请求频率）
func generateCaptcha(c *gin.Context) {
	type CaptchaRes struct {
		Id      string `json:"id"`
		Captcha string `json:"captcha"`
	}

	//ctx := context.Background()

	// **创建验证码**
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)

	// **生成验证码 ID 和 Base64 编码的图片**
	id, b64s, err := captcha.Generate()
	if err != nil {
		response.Success(c, response.ServerError, errors.New("生成验证码失败"))
		return
	}

	//// **存入 Redis，验证码 5 分钟有效**
	//captchaKey := fmt.Sprintf("captcha:%s", id)
	//err = config.R.Set(ctx, captchaKey, id, 5*time.Minute).Err()
	//if err != nil {
	//	response.Success(c, response.ServerError, errors.New("存储验证码失败"))
	//	return
	//}

	res := CaptchaRes{
		Id:      id,
		Captcha: b64s,
	}

	log.Println("✅ 生成验证码:", res)
	response.SuccessWithData(c, response.SuccessCode, res)
}
