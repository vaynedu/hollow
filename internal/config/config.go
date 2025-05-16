package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}

func NewConfig(cfgPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")

	// 监听配置文件改变
	v.WatchConfig()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return &Config{v}, nil
}

// 示例配置文件（example/config.yaml）
//# server:
//#   http:
//#     addr: ":8080"
//# log:
//#   level: "debug"
//#   file: "app.log"
