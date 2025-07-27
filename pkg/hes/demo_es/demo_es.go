package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

func main() {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200", // Elasticsearch 地址
		},
		// Username: "elastic",  // 用户名（如果启用了安全认证）
		// Password: "password", // 密码
	})
	if err != nil {
		log.Fatalf("创建 Elasticsearch 客户端失败: %s", err)
	}

	// 测试连接
	res, err := es.Info()
	if err != nil {
		log.Fatalf("连接 Elasticsearch 失败: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Elasticsearch 返回错误: %s", res.Status())
	}

	fmt.Println("成功连接到 Elasticsearch")

	// 创建索引请求
	req := esapi.IndexRequest{
		Index:      "products",                                  // 索引名
		DocumentID: "1111",                                      // 文档ID（可选，不指定时ES会自动生成）
		Body:       bytes.NewReader([]byte(`{"name": "1111"}`)), // 请求体
		Refresh:    "true",                                      // 插入后刷新索引，确保数据可立即搜索
	}

	// 执行请求
	res, err = req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("插入数据失败: %s", err)
	}
	defer res.Body.Close()

	// 处理响应
	if res.IsError() {
		log.Fatalf("Elasticsearch 返回错误: %s", res.Status())
	} else {
		// 解析响应
		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			log.Fatalf("解析响应失败: %s", err)
		}
		fmt.Printf("数据插入成功,ID: %s,版本: %d\n",
			result["_id"], int(result["_version"].(float64)))
	}
}
