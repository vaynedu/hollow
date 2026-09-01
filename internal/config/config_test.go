package config

import (
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
		So(config.Log.LogLevel, ShouldEqual, "debug")
	})
}
