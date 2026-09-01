// Package hredis 提供 github.com/redis/go-redis/v9 的薄封装:
//   - Config 结构体含连接池 / 超时 / DB / 密码,默认值兜底
//   - NewClient 构造后立即 Ping 一次验证连通性
//   - Ping 单独导出供外部探活
//
// 不重造 Redis 原生 API,获取到 *redis.Client 后直接调用 Get/Set/HSet 等。
package hredis
