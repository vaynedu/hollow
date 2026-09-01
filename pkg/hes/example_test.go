package hes

import (
	"context"
	"fmt"
	"time"
)

// Example 不带 // Output:,需要可用的 ES 实例,仅展示调用形态
func ExampleNewClient() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := NewClient(ctx, Config{
		Addresses: []string{"http://127.0.0.1:9200"},
		Username:  "elastic",
		Password:  "changeme",
	})
	if err != nil {
		fmt.Println("connect failed:", err)
		return
	}

	// 拿到原生 *elasticsearch.Client 后,可直接用 v8 的 esapi.*
	res, _ := client.Info(client.Info.WithContext(ctx))
	defer res.Body.Close()
	fmt.Println(res.Status())
}
