package hlark

import "fmt"

func ExampleGenSign() {
	// 固定 timestamp + secret 验证签名稳定性
	sign1, _ := GenSign(1700000000, "my-secret")
	sign2, _ := GenSign(1700000000, "my-secret")
	fmt.Println(sign1 == sign2)
	// Output:
	// true
}
