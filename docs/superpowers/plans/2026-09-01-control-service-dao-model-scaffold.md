# Hollow 分层脚手架 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 Hollow 新项目以一级目录 `control → service → dao → model` 组织手写代码，并将 Proto HTTP Adapter 保持为不可手改的生成代码。

**Architecture:** `protoc-gen-myhttp` 继续在 `proto/*_myhttp.pb.go` 中生成 Gin 路由、请求绑定和响应适配；`control` 实现生成的 Proto Service 接口并编排业务；`service`、`dao`、`model` 是不受 `make proto` 覆盖的手写层。Hollow 核心与 myhttp 插件不修改，只改 Hollow CLI 模板、测试、文档和现有 `life-core` 的装配。

**Tech Stack:** Go 1.25+、Cobra、`embed`/`text/template`、Protocol Buffers、Gin、Hollow

**Spec:** `docs/superpowers/specs/2026-09-01-control-service-dao-model-scaffold-design.md`

## Global Constraints

- 不新增项目级 `handler/` 或 `server/` 目录。
- `control` 使用 `context.Context` 和 Proto Request/Response，不依赖 `gin.Context`。
- `make proto` 只更新 `proto/*.pb.go` 与 `proto/*_myhttp.pb.go`。
- 手写目录 `control/`、`service/`、`dao/`、`model/` 只由 `hollow-cli init` 首次创建。
- 不引入 MySQL、GORM、Redis、IoC、Repository 接口或新的 CLI 命令。

---

### Task 1: 用测试定义新的生成项目结构

**Files:**
- Modify: `cmd/hollow_cli/generator/project_test.go`
- Modify: `cmd/hollow_cli/generator/integration_test.go`

**Interfaces:**
- Consumes: `InitProject(projectPath string, options ProjectOptions) error`
- Produces: 新目录树和覆盖保护的回归测试。

- [ ] **Step 1: 修改黄金目录测试**

将期望文件调整为：

```go
want := []string{
    ".gitignore",
    "Makefile",
    "README.md",
    "conf.yaml",
    "config/config.go",
    "control/control.go",
    "dao/dao.go",
    "go.mod",
    "main.go",
    "model/model.go",
    "proto/lifelog.proto",
    "router/router.go",
    "service/service.go",
}
```

- [ ] **Step 2: 修改模板内容断言**

断言 `control/control.go` 包含：

```go
type Control struct {
    proto.UnimplementedLifelogServiceService
}
func Get() *Control
```

断言 Router 包含 `control.Get()` 且不包含 `service.Get()`；断言 `service/service.go`、`dao/dao.go`、`model/model.go` 不依赖 Proto 或 Gin。

- [ ] **Step 3: 增加 make proto 不覆盖手写层测试**

在集成测试生成项目后，分别向四个手写文件追加唯一标记，执行 `make proto`，再逐字比较执行前后的文件内容。

- [ ] **Step 4: 运行测试并确认按预期失败**

Run:

```bash
go test ./cmd/hollow_cli/generator -count=1
```

Expected: FAIL，报告缺少 `control/dao/model`，Router 仍引用 `service.Get()`。

---

### Task 2: 实现分层项目模板

**Files:**
- Create: `cmd/hollow_cli/generator/templates/control.go.tmpl`
- Create: `cmd/hollow_cli/generator/templates/dao.go.tmpl`
- Create: `cmd/hollow_cli/generator/templates/model.go.tmpl`
- Modify: `cmd/hollow_cli/generator/templates/service.go.tmpl`
- Modify: `cmd/hollow_cli/generator/templates/router.go.tmpl`
- Modify: `cmd/hollow_cli/generator/render.go`
- Modify: `cmd/hollow_cli/generator/templates/README.md.tmpl`
- Modify: `README.md`

**Interfaces:**
- Consumes: `ProjectConfig.ModuleName`、`ProjectConfig.ServiceName`。
- Produces: `control.Get() *Control`，其 `Control` 嵌入 `proto.Unimplemented<Service>Service` 并满足生成 Adapter 所需接口。

- [ ] **Step 1: 新增 Control 模板**

```go
package control

import "{{.ModuleName}}/proto"

type Control struct {
    proto.Unimplemented{{.ServiceName}}Service
}

var instance = new(Control)

func Get() *Control { return instance }
```

- [ ] **Step 2: 将 Service、DAO、Model 改为空业务骨架**

三个文件分别只保留包声明和中文职责注释，不导入 Proto、Gin 或数据库库。

- [ ] **Step 3: 修改 Router 装配**

将 import 和注册参数从 `service.Get()` 改为 `control.Get()`。

- [ ] **Step 4: 注册新模板文件**

在 `projectFiles` 中加入：

```go
{path: "control/control.go", template: "control.go.tmpl"},
{path: "dao/dao.go", template: "dao.go.tmpl"},
{path: "model/model.go", template: "model.go.tmpl"},
```

保留 `service/service.go`，但使用纯业务模板。

- [ ] **Step 5: 更新项目说明**

在生成 README 和 Hollow 根 README 中写明自动 Adapter 与四层手写目录的职责、调用链和禁止覆盖规则。

- [ ] **Step 6: 运行单元测试并确认通过**

Run:

```bash
go test ./cmd/hollow_cli/generator -count=1
```

Expected: PASS。

- [ ] **Step 7: 运行集成测试**

Run:

```bash
go test -tags=integration ./cmd/hollow_cli/generator -count=1
```

Expected: PASS，生成项目可完成 Proto、依赖、测试和构建，且手写层未被覆盖。

- [ ] **Step 8: 提交 Hollow 实现**

```bash
git add cmd/hollow_cli/generator README.md
git commit -m "feat: scaffold layered Hollow projects"
```

---

### Task 3: 迁移现有 life-core

**Files:**
- Create: `/Users/nikki/go/src/life-core/control/control.go`
- Create: `/Users/nikki/go/src/life-core/dao/dao.go`
- Create: `/Users/nikki/go/src/life-core/model/model.go`
- Modify: `/Users/nikki/go/src/life-core/service/service.go`
- Modify: `/Users/nikki/go/src/life-core/router/router.go`
- Modify: `/Users/nikki/go/src/life-core/README.md`

**Interfaces:**
- Consumes: `proto.RegisterLifeCoreServiceGinRouter(*gin.Engine, proto.LifeCoreServiceService)`。
- Produces: `control.Get() *Control`，Router 使用该实例注册路由。

- [ ] **Step 1: 记录迁移前基线**

Run:

```bash
cd /Users/nikki/go/src/life-core
make test && make build
```

Expected: PASS。

- [ ] **Step 2: 创建一级目录和 Control 实现骨架**

创建与新 CLI 模板完全一致的 `control/control.go`、`dao/dao.go`、`model/model.go`。

- [ ] **Step 3: 转换 Service 与 Router**

Service 去除 Proto 接口嵌入和单例；Router 导入 `control` 并传入 `control.Get()`。

- [ ] **Step 4: 更新 README**

记录 `proto adapter → control → service → dao → model` 的开发流程。

- [ ] **Step 5: 重新生成并验证**

Run:

```bash
make proto PROTO_INCLUDE=/usr/local/include
make deps
make test
make build
```

Expected: PASS；访问 `/-/health` 返回 200，未实现的 `/v1/ping` 继续由统一响应返回 500。

---

### Task 4: 全量回归与收尾

**Files:**
- Verify only: Hollow 仓库与 `/Users/nikki/go/src/life-core`

**Interfaces:**
- Consumes: 前三项任务的最终代码。
- Produces: 可合并、可推送的验证结果。

- [ ] **Step 1: Hollow 全量测试**

```bash
go test ./... -count=1
go test -race ./... -count=1
```

- [ ] **Step 2: 从零生成临时项目**

构建最新 `hollow-cli`，在临时目录生成 `smoke-core`，执行：

```bash
make proto PROTO_INCLUDE=/usr/local/include
make deps
make test
make build
```

- [ ] **Step 3: 核对变更范围和身份**

```bash
git diff --check
git status --short
git log -1 --format='%an <%ae> | %cn <%ce>'
```

Expected: Hollow 提交身份为 `vaynedu <1219345363@qq.com>`，不存在无关改动。
