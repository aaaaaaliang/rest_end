package config

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// JWTConfig 令牌配置
type JWTConfig struct {
	Secret    string `mapstructure:"secret"`     // JWT 密钥
	ExpiresIn int    `mapstructure:"expires_in"` // 过期时间（秒）
}

// JWT 配置变量
var (
	JWTSecret     string
	JWTExpiration time.Duration
)

// InitJWT 初始化 JWT 配置
func InitJWT() {
	JWTSecret = G.JWT.Secret

	// **修正：正确解析 ExpiresIn**
	JWTExpiration = time.Duration(G.JWT.ExpiresIn) * time.Second
	if JWTExpiration <= 0 {
		JWTExpiration = 72 * time.Hour // 默认 72 小时
	}
}

// GenerateJWT 生成 JWT
func GenerateJWT(userCode string) (string, error) {
	claims := jwt.MapClaims{
		"user_code": userCode,
		"exp":       time.Now().Add(JWTExpiration).Unix(), // 过期时间
		"iat":       time.Now().Unix(),                    // 签发时间
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// ParseJWT 解析 JWT 并返回 `user_code`
func ParseJWT(tokenString string) (string, error) {
	// **解析 Token**
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return "", err
	}

	// **提取 `claims` 并验证**
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// **检查 Token 是否过期**
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				return "", errors.New("token 已过期")
			}
		}

		// **返回 `user_code`**
		if userCode, ok := claims["user_code"].(string); ok {
			return userCode, nil
		}
	}

	return "", errors.New("无效的 Token")
}
