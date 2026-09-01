// Package hes 提供 github.com/elastic/go-elasticsearch/v8 的薄封装:
//   - Config 配置(支持 Addresses/CloudID + Basic Auth/APIKey)
//   - NewClient 构造时自动 Info() 探活
//   - Ping 单独导出供外部探活
//
// 不重造 ES 原生 API,获取到 *elasticsearch.Client 后直接调用 v8 的
// esapi.* / Search() / Index() 等方法。
package hes
