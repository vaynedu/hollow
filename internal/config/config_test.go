package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// 创建临时配置文件
	configContent := `host: 127.0.0.1:8090`
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
	t.Logf("config: %v", config.GetString("host"))
}
