// Package hscheduler 提供轻量级的进程内定时任务调度。
//
// 简单任务可通过 RegisterFunc 直接注册闭包，复杂任务通过 Job 描述名称、
// Cron 表达式和执行逻辑，再将调度器接入
// Hollow 的 Startup、Shutdown 生命周期。调度器负责单次超时、防止任务重叠、
// 启动补跑、panic 恢复、执行日志和优雅关闭，不感知业务配置及依赖组装。
package hscheduler
