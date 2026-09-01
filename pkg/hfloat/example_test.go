package hfloat

import "fmt"

func ExampleAddFloat() {
	// 经典浮点精度问题:0.1 + 0.2 用 float 直接相加得 0.30000000000000004
	r := AddFloat(0.1, 0.2)
	fmt.Println(r)
	// Output:
	// 0.3
}

func ExampleAddStringFloat() {
	r, _ := AddStringFloat("0.1", "0.2")
	fmt.Println(r)
	// Output:
	// 0.3
}
