package config

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfig(t *testing.T) {
	Convey("NewConfig 读取临时 yaml 配置", t, func() {
		configContent := `host: 127.0.0.1:8090
log:
  level: debug
db:
  dsn: root:123456@tcp(127.0.0.1:3306)/marketing?charset=utf8mb4&parseTime=True&loc=Local
  dialect: mysql
redis:
  addr: 127.0.0.1:6379
  password: 123456
  db: 0
`
		configFile, err := os.CreateTemp(".", "test_config.yaml")
		So(err, ShouldBeNil)
		defer os.Remove(configFile.Name())

		_, err = configFile.WriteString(configContent)
		So(err, ShouldBeNil)

		// 关闭文件以确保内容写入磁盘
		So(configFile.Close(), ShouldBeNil)

		// 测试 NewConfig 函数
		config, err := NewConfig(".", configFile.Name())
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		t.Logf("config host: %v", config.GetString("host"))
		t.Logf("config log: %v", config.Log)
		t.Logf("config db: %v", config.Db)
		t.Logf("config redis: %v", config.Redis)
	})
}
