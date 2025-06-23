package hexcel

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type RowS struct {
	Name              string          `json:"name"`
	CurrencyPriceList []CurrencyPrice `json:"currency_price_list"`
}

type CurrencyPrice struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

// buildRowsFromRecords 从 CSV 记录中构建 RowS 结构
func buildRowsFromRecords(records [][]string) ([]RowS, error) {
	if len(records) < 2 {
		return nil, errors.New("not enough data to build rows")
	}

	var rows []RowS
	currencies := records[1] // 第二行为币种

	for _, record := range records[2:] { // 从第三行开始是金额
		if len(record) == 0 {
			continue
		}
		name := fmt.Sprintf("product_price_cny_%s", record[0])

		var currencyPriceList []CurrencyPrice
		for i, currency := range currencies {
			if i < len(record) {
				currencyPriceList = append(currencyPriceList, CurrencyPrice{
					Currency: currency,
					Amount:   record[i],
				})
			}
		}

		rows = append(rows, RowS{
			Name:              name,
			CurrencyPriceList: currencyPriceList,
		})
	}

	return rows, nil
}

// generateSQLStatements 生成插入 SQL 语句
func generateSQLStatements(rows []RowS, groupID int, createdBy, deletedBy, updatedBy string) ([]string, error) {
	var sqls []string

	for _, row := range rows {
		jsonBytes, err := json.Marshal(row.CurrencyPriceList)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal currency_price_list for %s: %w", row.Name, err)
		}
		jsonStr := string(jsonBytes)

		sql := fmt.Sprintf(
			`INSERT INTO price (id, name, currency_price_list, created_by, deleted_by, updated_by) VALUES (%d, "%s", '%s', '%s', '%s', '%s');`,
			groupID, row.Name, jsonStr, createdBy, deletedBy, updatedBy,
		)
		sqls = append(sqls, sql)
	}

	return sqls, nil
}

// writeSQLToFile 将 SQL 语句写入文件（覆盖模式）
func writeSQLToFile(filename string, sqls []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for _, sql := range sqls {
		_, err := file.WriteString(sql + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}

// exportToSQLFile 主流程：读取 CSV -> 构建结构 -> 生成 SQL -> 写入文件
func exportToSQLFile(csvFilename, sqlFilename string, groupID int, createdBy, deletedBy, updatedBy string) error {
	records, err := GetCSVData(csvFilename)
	if err != nil {
		return err
	}

	//  CNY USD TWD
	//   1   2   3
	rows, err := buildRowsFromRecords(records)
	if err != nil {
		return err
	}

	sqls, err := generateSQLStatements(rows, groupID, createdBy, deletedBy, updatedBy)
	if err != nil {
		return err
	}

	err = writeSQLToFile(sqlFilename, sqls)
	if err != nil {
		return err
	}

	fmt.Println("SQL export completed successfully.")
	return nil
}
