package config

// CorsConfig 结构体映射 CORS 配置
type CorsConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`     // 允许的前端域名
	AllowMethods     []string `mapstructure:"allow_methods"`     // 允许的 HTTP 方法
	AllowHeaders     []string `mapstructure:"allow_headers"`     // 允许的 HTTP 头部
	AllowCredentials bool     `mapstructure:"allow_credentials"` // 是否允许携带 Cookie
	MaxAge           int      `mapstructure:"max_age"`           // 预检请求的缓存时间（秒）
}
