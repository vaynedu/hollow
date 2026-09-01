package hredis

import (
	"context"
	"fmt"
	"time"
)

// 注意:此 Example 不带 // Output:,因为它需要可用的 Redis 实例,
// 仅作调用形态展示。
func ExampleNewClient() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := NewClient(ctx, Config{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	if err != nil {
		fmt.Println("connect failed:", err)
		return
	}
	defer client.Close()

	// 直接使用原生 API
	_ = client.Set(ctx, "key", "value", 0).Err()
}
