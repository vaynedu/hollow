package config

import (
	"time"

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

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
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

	return &config, nil
}
