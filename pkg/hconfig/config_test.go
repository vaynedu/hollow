package hconfig

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var errInvalidName = errors.New("name 不合法")

type testConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

type testSchedulerConfig struct {
	Scheduler struct {
		Enabled  bool   `mapstructure:"enabled"`
		Timezone string `mapstructure:"timezone"`
	} `mapstructure:"scheduler"`
}

func (c *testConfig) Validate() error {
	if c.Name == "invalid" {
		return errInvalidName
	}
	return nil
}

func TestLoad(t *testing.T) {
	Convey("Load 读取本地 YAML", t, func() {
		dir := t.TempDir()
		writeConfig := func(content string) {
			So(os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte(content), 0o600), ShouldBeNil)
		}

		Convey("YAML 覆盖已有字段并保留未配置的默认值", func() {
			writeConfig("name: life-core\n")
			cfg := &testConfig{Port: 8080}

			err := Load(dir, "conf", cfg)

			So(err, ShouldBeNil)
			So(cfg.Name, ShouldEqual, "life-core")
			So(cfg.Port, ShouldEqual, 8080)
		})

		Convey("拒绝 nil 和非指针目标", func() {
			writeConfig("name: life-core\n")
			var nilConfig *testConfig

			So(errors.Is(Load(dir, "conf", nil), ErrInvalidTarget), ShouldBeTrue)
			So(errors.Is(Load(dir, "conf", testConfig{}), ErrInvalidTarget), ShouldBeTrue)
			So(errors.Is(Load(dir, "conf", nilConfig), ErrInvalidTarget), ShouldBeTrue)
		})

		Convey("配置文件不存在时错误包含文件名", func() {
			err := Load(dir, "missing", &testConfig{})

			So(err, ShouldNotBeNil)
			So(strings.Contains(err.Error(), "missing.yaml"), ShouldBeTrue)
		})

		Convey("反序列化后执行配置校验", func() {
			writeConfig("name: invalid\n")

			err := Load(dir, "conf", &testConfig{})

			So(errors.Is(err, errInvalidName), ShouldBeTrue)
		})
	})
}

func TestLoadWithEnvOverridesYAML(t *testing.T) {
	Convey("LoadWithEnv 支持环境变量", t, func() {
		dir := t.TempDir()
		t.Setenv("LIFE_CORE_SCHEDULER_ENABLED", "true")
		t.Setenv("LIFE_CORE_SCHEDULER_TIMEZONE", "Asia/Shanghai")

		Convey("覆盖 YAML 中的配置", func() {
			content := "scheduler:\n  enabled: false\n  timezone: UTC\n"
			So(os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte(content), 0o600), ShouldBeNil)
			var cfg testSchedulerConfig

			err := LoadWithEnv(dir, "conf", "LIFE_CORE", &cfg)

			So(err, ShouldBeNil)
			So(cfg.Scheduler.Enabled, ShouldBeTrue)
			So(cfg.Scheduler.Timezone, ShouldEqual, "Asia/Shanghai")
		})

		Convey("支持仅由环境变量提供的配置", func() {
			So(os.WriteFile(filepath.Join(dir, "conf.yaml"), []byte("{}\n"), 0o600), ShouldBeNil)
			var cfg testSchedulerConfig

			err := LoadWithEnv(dir, "conf", "LIFE_CORE", &cfg)

			So(err, ShouldBeNil)
			So(cfg.Scheduler.Enabled, ShouldBeTrue)
			So(cfg.Scheduler.Timezone, ShouldEqual, "Asia/Shanghai")
		})
	})
}
