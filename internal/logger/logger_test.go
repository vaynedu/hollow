package logger

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vaynedu/hollow/internal/config"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	Convey("InitLogger 各种配置", t, func() {
		Convey("default config (nil)", func() {
			log, err := InitLogger(nil)
			So(err, ShouldBeNil)
			So(log, ShouldNotBeNil)
		})

		Convey("console config", func() {
			cfg := &config.Config{
				Log: config.LogConfig{
					LogLevel:   "info",
					OutputMode: "console",
				},
			}
			log, err := InitLogger(cfg)
			So(err, ShouldBeNil)
			So(log, ShouldNotBeNil)
		})

		Convey("file config", func() {
			cfg := &config.Config{
				Log: config.LogConfig{
					LogLevel:    "debug",
					OutputMode:  "file",
					LogFileName: "test.log",
					MaxSize:     10,
					MaxAge:      7,
				},
			}
			log, err := InitLogger(cfg)
			So(err, ShouldBeNil)
			So(log, ShouldNotBeNil)
		})
	})
}

func TestGetLogger(t *testing.T) {
	Convey("GetLogger 在 logger 为 nil 时也应返回有效 logger", t, func() {
		// 重置全局 logger 为 nil 后再获取
		logger = nil
		log := GetLogger()
		So(log, ShouldNotBeNil)
	})
}

func TestLogFunctions(t *testing.T) {
	Convey("各级别格式化 / 结构化日志函数可调用不 panic", t, func() {
		_, err := InitLogger(nil)
		So(err, ShouldBeNil)

		// 格式化日志函数
		Debugf("test debug: %s", "message")
		Infof("test info: %s", "message")
		Warnf("test warn: %s", "message")
		Errorf("test error: %s", "message")

		// 带字段日志函数
		Debug("debug message", zap.String("key", "value"))
		Info("info message", zap.String("key", "value"))
		Warn("warn message", zap.String("key", "value"))
		Error("error message", zap.String("key", "value"))
	})
}

func TestWithFields(t *testing.T) {
	Convey("WithFields 返回带字段的 logger", t, func() {
		_, err := InitLogger(nil)
		So(err, ShouldBeNil)

		log := WithFields(zap.String("request_id", "12345"))
		So(log, ShouldNotBeNil)
		log.Info("test with fields")
	})
}
