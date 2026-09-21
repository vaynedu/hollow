package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfig(t *testing.T) {
	Convey("NewConfig 读取临时 yaml 配置", t, func() {
		configContent := "server:\n" +
			"  host: 127.0.0.1:8090\n" +
			"  shutdown_timeout: 5s\n" +
			"log:\n" +
			"  level: debug\n"
		dir := t.TempDir()
		So(os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte(configContent), 0o600), ShouldBeNil)

		config, err := NewConfig(dir, "conf")
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.Server.Host, ShouldEqual, "127.0.0.1:8090")
		So(config.Server.ShutdownTimeout, ShouldEqual, 5*time.Second)
		So(config.Log.Level, ShouldEqual, "debug")
	})
}

func TestNewConfigDefaultsAndValidation(t *testing.T) {
	Convey("NewConfig 应用默认值并校验服务配置", t, func() {
		dir := t.TempDir()

		Convey("空配置使用可运行的服务默认值", func() {
			So(os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte("{}\n"), 0o600), ShouldBeNil)

			cfg, err := NewConfig(dir, "conf")

			So(err, ShouldBeNil)
			So(cfg.Server.Host, ShouldEqual, "127.0.0.1:8080")
			So(cfg.Server.ShutdownTimeout, ShouldEqual, 10*time.Second)
		})

		Convey("拒绝非正数关闭超时", func() {
			content := "server:\n  host: 127.0.0.1:8080\n  shutdown_timeout: -1s\n"
			So(os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte(content), 0o600), ShouldBeNil)

			_, err := NewConfig(dir, "conf")

			So(errors.Is(err, ErrInvalidShutdownTimeout), ShouldBeTrue)
		})
	})
}

func TestNewConfigWithEnvOverridesYAML(t *testing.T) {
	Convey("NewConfigWithEnv 使用环境变量覆盖核心配置", t, func() {
		dir := t.TempDir()
		content := "server:\n  host: 127.0.0.1:8090\n  shutdown_timeout: 5s\nlog:\n  level: info\n"
		So(os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte(content), 0o600), ShouldBeNil)
		t.Setenv("LIFELOG_SERVER_HOST", "127.0.0.1:9090")
		t.Setenv("LIFELOG_LOG_LEVEL", "debug")

		cfg, err := NewConfigWithEnv(dir, "conf", "LIFELOG")

		So(err, ShouldBeNil)
		So(cfg.Server.Host, ShouldEqual, "127.0.0.1:9090")
		So(cfg.Log.Level, ShouldEqual, "debug")
	})
}
