package hecode

import "fmt"

func ExampleNew() {
	err := New(9001, "user not found")
	fmt.Println(Code(err))
	fmt.Println(err)
	// Output:
	// 9001
	// code=9001, msg=user not found
}

func ExampleWrap() {
	base := New(9002, "db query failed")
	wrapped := Wrap(base, "load user profile")
	fmt.Println(Code(wrapped))          // 保留原 code
	fmt.Println(Cause(wrapped) == base) // 链尾就是 base
	// Output:
	// 9002
	// true
}
