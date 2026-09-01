package hidgenerator

import "fmt"

func ExampleNewUuid() {
	gen := NewUuid()
	id := gen.GenerateRequestID()
	fmt.Println(len(id)) // UUID v4 标准长度 36
	// Output:
	// 36
}

func ExampleNewSnowflake() {
	gen, _ := NewSnowflake(1)
	id := gen.GenerateInt64()
	fmt.Println(id > 0)
	// Output:
	// true
}
