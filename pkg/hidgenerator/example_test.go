package hidgenerator

import "fmt"

func ExampleNewUuid() {
	gen := NewUuid()
	id := gen.GenerateRequestID()
	fmt.Println(len(id)) // UUID v4 标准长度 36
	// Output:
	// 36
}
