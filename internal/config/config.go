package config

import (
	"github.com/spf13/viper"
)

// LogConfig 定义日志配置结构体
type LogConfig struct {
	LogLevel    string `mapstructure:"level"`
	OutputMode  string `mapstructure:"output_mode"`
	LogFileName string `mapstructure:"file"`
	MaxSize     int    `mapstructure:"max_size"`
	MaxAge      int    `mapstructure:"max_age"`
}

type Config struct {
	*viper.Viper
	Host  string      `mapstructure:"host"`
	Log   LogConfig   `mapstructure:"log"`
	Db    DbConfig    `mapstructure:"db"`
	Redis RedisConfig `mapstructure:"redis"`
}

func NewConfig(path string, configFileName string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(configFileName)
	v.SetConfigType("yaml")

	// 监听配置文件改变  // todo 这里需要验证
	// v.WatchConfig()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}
	config.Viper = v // 假设 Config 结构体有 Viper 字段，添加赋值操作

	return &config, nil
}

// 示例配置文件（example/conf.yaml）
//# server:
//#   http:
//#     addr: ":8080"
//# log:
//#   level: "debug"
//#   output_mode: "console"
//#   file: "app.log"
//#   max_size: 100
//#   max_age: 30
