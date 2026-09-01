package hcsv

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExampleGetCSVData() {
	dir, _ := os.MkdirTemp("", "hcsv-example-")
	defer os.RemoveAll(dir)
	f := filepath.Join(dir, "data.csv")
	_ = os.WriteFile(f, []byte("name,age\nAlice,30\nBob,25"), 0644)

	rows, _ := GetCSVData(f)
	for _, row := range rows {
		fmt.Println(row)
	}
	// Output:
	// [name age]
	// [Alice 30]
	// [Bob 25]
}
