package config

// AppConfig 应用配置
type AppConfig struct {
	Name  string `mapstructure:"name"`
	Host  string `mapstructure:"host"`
	Port  int    `mapstructure:"port"`
	Env   string `mapstructure:"env"`
	Debug bool   `mapstructure:"debug"`
}
