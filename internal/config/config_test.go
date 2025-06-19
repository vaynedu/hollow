package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
    // 创建临时配置文件
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
	if err != nil {
		t.Fatalf("创建临时配置文件失败: %v", err)
	}
	defer os.Remove(configFile.Name())

	_, err = configFile.WriteString(configContent)
	if err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	// 关闭文件以确保内容写入磁盘
	configFile.Close()

	// 测试 NewConfig 函数
	config, err := NewConfig(".", configFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)
	t.Logf("config host: %v", config.GetString("host"))
	// 打印yaml结构体
	t.Logf("config log: %v", config.Log)
	// 打印db结构体
	t.Logf("config db: %v", config.Db)
	// 打印redis结构体
	t.Logf("config redis: %v", config.Redis)

}
