package logger

import (
	"github.com/vaynedu/hollow/internal/config"
	"github.com/vaynedu/hollow/pkg/hlog"
	"go.uber.org/zap"
)

// InitLogger 创建日志实例。
func InitLogger(cfg *config.Config) (*zap.Logger, error) {
	var logConfig hlog.Config
	if cfg != nil {
		logConfig = cfg.Log
	}
	return hlog.New(logConfig)
}
