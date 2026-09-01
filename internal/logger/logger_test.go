package logger

import (
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vaynedu/hollow/internal/config"
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
					Level:      "info",
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
					Level:      "debug",
					OutputMode: "file",
					File:       filepath.Join(t.TempDir(), "test.log"),
					MaxSize:    10,
					MaxAge:     7,
				},
			}
			log, err := InitLogger(cfg)
			So(err, ShouldBeNil)
			So(log, ShouldNotBeNil)
		})
	})
}
