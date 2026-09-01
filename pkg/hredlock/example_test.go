package hredlock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Example 不带 // Output:,需要可用的 Redis 实例,仅展示调用形态
func ExampleRedlock_Lock() {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	defer client.Close()

	rl, err := NewRedlock(client)
	if err != nil {
		fmt.Println("init failed:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 加锁,过期时间 5s
	mu, err := rl.Lock(ctx, "order:1001", 5*time.Second)
	if err != nil {
		fmt.Println("lock failed:", err)
		return
	}
	defer func() {
		if err := mu.Unlock(context.Background()); err != nil {
			fmt.Println("unlock failed:", err)
		}
	}()

	// 临界区...
}
