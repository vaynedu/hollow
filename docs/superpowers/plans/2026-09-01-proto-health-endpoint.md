# Proto Health Endpoint Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 删除 Hollow 内置健康路由，并让所有新项目自带经过 Proto、Control、Service 的 `/v1/health` 接口，同时完善服务不可用和超时的 HTTP 状态映射。

**Architecture:** Hollow 只提供 Web 生命周期、统一响应和错误码映射，不再拥有服务级健康路由。脚手架生成 Health Proto、Control 实现和 Service 实现；HTTP Adapter 仍由 `protoc-gen-myhttp` 生成，手写实现不会被 `make proto` 覆盖。

**Tech Stack:** Go、Gin、Protocol Buffers、Hollow CLI 模板、`pkg/hecode`

**Spec:** 用户确认的双层职责调整：移除 `/-/health`，默认生成 `/v1/health`。

## Global Constraints

- Health 必须在 Proto 中定义为 GET `/v1/health`。
- 调用链必须是生成 Adapter → `control.Health` → `service.Health`。
- Control 依赖 Proto，Service 不依赖 Proto 或 Gin。
- `make proto` 不得覆盖 `control/health.go` 或 `service/health.go`。
- `ErrService` 映射 HTTP 503，`ErrTimeout` 映射 HTTP 504。

### Task 1: 删除框架内置健康路由

- [ ] 修改现有测试，断言 `NewApp` 不注册 `/-/health`。
- [ ] 将 Server 测试改为显式注册测试路由。
- [ ] 运行测试确认失败。
- [ ] 删除 `hollow.go` 中的内置路由并确认测试通过。

### Task 2: 完善错误码到 HTTP 状态映射

- [ ] 为 `ErrService` 503、`ErrTimeout` 504 添加失败测试。
- [ ] 修改 `pkg/hecode.Resolve` 的映射。
- [ ] 运行 `pkg/hecode` 和中间件测试。

### Task 3: 默认生成完整 Health 调用链

- [ ] 修改脚手架测试，期望 Health Proto、`control/health.go`、`service/health.go`。
- [ ] 修改集成测试，验证两个 Health 文件不被 `make proto` 覆盖。
- [ ] 运行测试确认失败。
- [ ] 修改 Proto 模板并新增 Health Control、Service 模板。
- [ ] 注册新模板并更新 README。
- [ ] 运行 CLI 单元与集成测试。

### Task 4: 迁移 life-core 并全量验证

- [ ] 将 `life-core` 的 Ping Proto 改为 Health。
- [ ] 添加 `control.Health` 和 `service.Health`。
- [ ] 执行 `make proto`、`make deps`、`make test`、`make build`。
- [ ] 启动服务并验证 `/v1/health` 返回 HTTP 200 和统一成功响应。
- [ ] 执行 Hollow 全量测试和 Race 测试。
