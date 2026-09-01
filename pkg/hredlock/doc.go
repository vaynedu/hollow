// Package hredlock 提供基于 github.com/go-redsync/redsync/v4 的分布式锁封装。
//
// 设计:
//   - Redlock 工厂接受业务已建立的 *redis.Client,不重复创建连接
//   - Lock 返回 *Mutex,*Mutex.Unlock 释放(典型 defer 调用)
//   - 锁 expiry 必须显式传入,避免框架默认值掩盖业务超时风险
//
// 不暴露 redsync 原生选项(WithTries / WithRetryDelay 等);
// 后续按需扩展。
package hredlock
