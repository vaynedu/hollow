// Package hidgenerator 提供请求 ID / 业务 ID 生成器接口与实现。
//
// IdGenerator 接口允许业务按需切换底层实现。
// 当前内置 UUID v4 一种实现;后续可扩展 snowflake / nanoid 等。
package hidgenerator
