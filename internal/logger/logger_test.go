package logger_test

import (
	"testing"

	"github.com/vaynedu/hollow/internal/config"
	"github.com/vaynedu/hollow/internal/logger"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	// 测试默认配置
	defaultCfg := &config.Config{}
	defaultLog, err := logger.InitLogger(defaultCfg)
	if err != nil {
		t.Errorf("使用默认配置初始化日志失败: %v", err)
	}
	if defaultLog == nil {
		t.Error("使用默认配置返回的日志实例为 nil")
	}
	defer defaultLog.Sync()

	defaultLog.Info("defaultLog测试默认info配置日志", zap.String("key", "value"))
	defaultLog.Warn("defaultLog测试默认warn配置日志", zap.String("key", "value"))
	defaultLog.Error("defaultLog测试默认error配置日志", zap.String("key", "value"))
	defaultLog.Debug("defaultLog测试默认debug配置日志", zap.String("key", "value"))

	// 测试文件输出模式
	fileCfg := &config.Config{
		Log: config.LogConfig{
			LogLevel:    "debug",
			OutputMode:  "file",
			LogFileName: "test.log",
			MaxSize:     50,
			MaxAge:      15,
		},
	}
	fileLog, err := logger.InitLogger(fileCfg)
	if err != nil {
		t.Errorf("使用文件输出模式初始化日志失败: %v", err)
	}
	if fileLog == nil {
		t.Error("使用文件输出模式返回的日志实例为 nil")
	}
	defer fileLog.Sync()
	fileLog.Info("fileLog测试文件info配置日志", zap.String("key", "value"))
	fileLog.Warn("fileLog测试文件warn配置日志", zap.String("key", "value"))
	fileLog.Error("fileLog测试文件error配置日志", zap.String("key", "value"))
	fileLog.Debug("fileLog测试文件debug配置日志", zap.String("key", "value"))

	// 测试控制台输出模式
	consoleCfg := &config.Config{
		Log: config.LogConfig{
			LogLevel:   "warn",
			OutputMode: "console",
		},
	}
	consoleLog, err := logger.InitLogger(consoleCfg)
	if err != nil {
		t.Errorf("使用控制台输出模式初始化日志失败: %v", err)
	}
	if consoleLog == nil {
		t.Error("使用控制台输出模式返回的日志实例为 nil")
	}
	defer consoleLog.Sync()
	consoleLog.Info("consoleLog测试控制台info配置日志", zap.String("key", "value"))
	consoleLog.Warn("consoleLog测试控制台warn配置日志", zap.String("key", "value"))
	consoleLog.Error("consoleLog测试控制台error配置日志", zap.String("key", "value"))
	consoleLog.Debug("consoleLog测试控制台debug配置日志", zap.String("key", "value"))
}
