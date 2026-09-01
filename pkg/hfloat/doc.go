// Package hfloat 提供基于 shopspring/decimal 的浮点运算工具。
//
// 适用于价格 / 金额 / 会计场景:
//   - 加减乘除走 decimal 避免精度丢失
//   - 比较走容差(默认 1e-6)避免直接 == 浮点判断
//   - 字符串与 float64 互转保留有效精度
//
// 不适用于追求极致性能的科学计算(decimal 慢于原生 float)。
package hfloat
