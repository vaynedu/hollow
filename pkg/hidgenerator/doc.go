// Package hidgenerator 提供请求 ID / 业务 ID 生成器接口与实现。
//
// IdGenerator 接口允许业务按需切换底层实现。
// 当前内置两种实现:
//   - Uuid:基于 google/uuid v4,无状态,适合 request_id / trace_id
//   - Snowflake:64bit Twitter 标准实现,有状态(需 machineID),
//     单机峰值 ~4M/s,适合业务主键 ID
package hidgenerator
