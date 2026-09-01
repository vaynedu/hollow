// Package hecode 提供带错码的错误体系与统一响应辅助。
//
// 错码段约定(系统预留):
//   - [1000,1100) 系统错(internal/network/timeout/...)
//   - [1100,1200) 参数错
//   - [1200,1300) 业务错(not found/forbidden/...)
//   - [1300,1400) 数据校验错
//   - [1400,1500) 数据库错
//   - [1500,1600) Redis 错
//
// 业务方建议从 9000 起定义自己的错码,避免与系统段冲突。
// 错码全局唯一,重复注册会 panic(在 init 阶段就能发现)。
//
// Wrap / WithMessage 对非 EcodeError 类型的 err 会用 ErrCodeUnknown(1099)
// 作为兜底错码;EcodeError 类型的 err 则保留原 code。
package hecode
