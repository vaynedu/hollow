package hlog

import (
	"errors"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("New 构造结构化 Logger", t, func() {
		Convey("零值配置使用安全默认值", func() {
			logger, err := New(Config{})

			So(err, ShouldBeNil)
			So(logger, ShouldNotBeNil)
		})

		Convey("支持文件输出和轮转配置", func() {
			logger, err := New(Config{
				Level:      "info",
				OutputMode: "file",
				File:       filepath.Join(t.TempDir(), "app.log"),
				MaxSize:    10,
				MaxBackups: 2,
				MaxAge:     7,
				Compress:   true,
			})

			So(err, ShouldBeNil)
			logger.Info("file logger")
			So(logger.Sync(), ShouldBeNil)
		})

		Convey("拒绝非法日志级别", func() {
			_, err := New(Config{Level: "verbose"})

			So(errors.Is(err, ErrInvalidLevel), ShouldBeTrue)
		})

		Convey("拒绝非法输出模式", func() {
			_, err := New(Config{OutputMode: "network"})

			So(errors.Is(err, ErrInvalidOutputMode), ShouldBeTrue)
		})
	})
}
