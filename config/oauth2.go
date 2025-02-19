package config

type Oauth2Config struct {
	ClientID     string `mapstructure:"client_id"`     // GitHub Client ID
	ClientSecret string `mapstructure:"client_secret"` // GitHub Client Secret
	RedirectURI  string `mapstructure:"redirect_uri"`  // OAuth2 回调 URL
	Scope        string `mapstructure:"scope"`         // 请求的权限范围
}
