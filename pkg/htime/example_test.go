package htime

import "fmt"

func ExampleParseTimeStamp() {
	t, _ := ParseTimeStamp("1672531200")
	fmt.Println(t.UTC().Format(TimeFormatDateTimeStandard))
	// Output:
	// 2023-01-01 00:00:00
}

func ExampleGetTotalDaysInMonth() {
	fmt.Println(GetTotalDaysInMonth(2024, 2)) // 闰年 2 月
	fmt.Println(GetTotalDaysInMonth(2023, 2)) // 平年 2 月
	// Output:
	// 29
	// 28
}
