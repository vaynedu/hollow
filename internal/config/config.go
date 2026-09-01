package config

import (
	"errors"
	"time"

	"github.com/vaynedu/hollow/pkg/hconfig"
	"github.com/vaynedu/hollow/pkg/hlog"
)

// LogConfig 保留原有名称，实际配置由公开的 hlog.Config 定义。
type LogConfig = hlog.Config

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
}

// ErrInvalidShutdownTimeout 表示优雅关闭超时配置无效。
var ErrInvalidShutdownTimeout = errors.New("config: server.shutdown_timeout must be greater than zero")

// Validate 校验 Hollow 核心配置。
func (c *Config) Validate() error {
	if c.Server.ShutdownTimeout <= 0 {
		return ErrInvalidShutdownTimeout
	}
	return nil
}

func NewConfig(path string, configFileName string) (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:            "127.0.0.1:8080",
			ShutdownTimeout: 10 * time.Second,
		},
	}
	if err := hconfig.Load(path, configFileName, config); err != nil {
		return nil, err
	}

	return config, nil
}
