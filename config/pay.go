package config

type PayConfig struct {
	PrivateKey   string `mapstructure:"private_key"`
	SetReturnUrl string `mapstructure:"set_return_url"`
	SetNotifyUrl string `mapstructure:"set_notify_url"`
	AppId        string `mapstructure:"app_id"`
	IsProd       bool   `mapstructure:"is_prod"`
}
