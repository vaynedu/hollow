// Package hresty 提供基于 github.com/go-resty/resty/v2 的 HTTP 客户端封装。
//
// 主要能力:
//   - NewRestyClient 带默认连接池 / 超时 / DualStack 的 transport
//   - GetTraceInfo 提取结构化请求追踪(DNS / TCP / TLS / Server time 等)
//
// 不打日志,由调用方负责把 RequestTrace 喂给自己的 logger;
// 不引入 internal/logger 依赖。
package hresty
