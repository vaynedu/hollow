package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/vaynedu/hollow/pkg/hes"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 通过 hes 创建 v8 客户端(NewClient 内部 Info() 验证连通)
	es, err := hes.NewClient(ctx, hes.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		log.Fatalf("创建 Elasticsearch 客户端失败: %s", err)
	}

	fmt.Println("成功连接到 Elasticsearch")

	// 拿到原生 *elasticsearch.Client 后,用 v8 esapi 调用
	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: "1111",
		Body:       bytes.NewReader([]byte(`{"name": "1111"}`)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es)
	if err != nil {
		log.Fatalf("插入数据失败: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Elasticsearch 返回错误: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Fatalf("解析响应失败: %s", err)
	}
	fmt.Printf("数据插入成功,ID: %s,版本: %d\n",
		result["_id"], int(result["_version"].(float64)))
}
