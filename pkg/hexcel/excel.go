package hexcel

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
)

func GetCSVData(filename string) ([][]string, error) {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file error: %w", err)
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv file error: %w", err)
	}

	if len(records) == 0 {
		return nil, errors.New("csv file is empty")
	}

	return records, nil
}
