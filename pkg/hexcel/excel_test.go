package hexcel

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetCSVData(t *testing.T) {
	// 创建临时目录
	 tempDir := t.TempDir()
	 testFile := filepath.Join(tempDir, "test.csv")

	 t.Run("正常读取CSV文件", func(t *testing.T) {
		// 创建测试CSV文件
		content := "name,age\nAlice,30\nBob,25"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		// 调用测试函数
		records, err := GetCSVData(testFile)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(records))
		assert.Equal(t, []string{"name", "age"}, records[0])
		assert.Equal(t, []string{"Alice", "30"}, records[1])
		assert.Equal(t, []string{"Bob", "25"}, records[2])
	})

	 t.Run("文件不存在错误", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.csv")
		records, err := GetCSVData(nonExistentFile)
		assert.Error(t, err)
		assert.Nil(t, records)
		assert.Contains(t, err.Error(), "open file error")
	})

	 t.Run("空文件错误", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.csv")
		if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
			t.Fatalf("创建空文件失败: %v", err)
		}

		records, err := GetCSVData(emptyFile)
		assert.Error(t, err)
		assert.Nil(t, records)
		assert.Equal(t, "csv file is empty", err.Error())
	})

	 t.Run("CSV格式错误", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.csv")
		// 写入格式错误的CSV（缺少引号闭合）
		content := "name,age\nAlice,30\nBob,25,"
		if err := os.WriteFile(invalidFile, []byte(content), 0644); err != nil {
			t.Fatalf("创建无效CSV文件失败: %v", err)
		}

		records, err := GetCSVData(invalidFile)
		assert.Error(t, err)
		assert.Nil(t, records)
		assert.Contains(t, err.Error(), "read csv file error")
	})
}