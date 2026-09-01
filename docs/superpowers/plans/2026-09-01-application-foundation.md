# Hollow Application Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Hollow 和 CLI 补齐本地类型化配置、上下文结构化日志、MySQL/Redis 初始化模板，并将 `life-core` 同步为可直接开发业务的参考项目。

**Architecture:** Hollow 只提供可复用基础能力，CLI 在业务项目中生成配置快照和数据库生命周期代码。请求日志通过 `context.Context` 贯穿 Control、Service、DAO；数据库依赖通过构造函数显式传递；Cache 保持可选。

**Tech Stack:** Go 1.25、Gin、Viper、Zap、GORM、MySQL Driver、go-redis v9、GoConvey

**Spec:** `docs/superpowers/specs/2026-09-01-application-foundation-design.md`

## Global Constraints

- 配置仅从本地 `conf.yaml` 启动加载一次。
- 不实现热更新、远程配置和环境变量覆盖。
- 不引入 IoC、服务发现、gRPC 或 Worker。
- MySQL、Redis 模板默认生成，通过 `enabled` 控制，默认关闭。
- `cache` 不进入固定生成目录。
- Service 不保存请求级可变状态。
- 不修改 `protoc-gen-myhttp`。
- 所有新增单元测试使用 GoConvey。
- 代码注释使用中文。

---

### Task 1: 本地类型化配置加载器

**Files:**
- Create: `pkg/hconfig/config.go`
- Create: `pkg/hconfig/config_test.go`
- Create: `pkg/hconfig/doc.go`
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Interfaces:**
- Produces: `type Validator interface { Validate() error }`
- Produces: `func Load(path, name string, target any) error`
- Consumes: Viper YAML reader

- [ ] **Step 1: 写失败测试**

在 `pkg/hconfig/config_test.go` 使用 GoConvey 覆盖：预置默认值被保留、YAML 字段覆盖默认值、nil/非指针目标拒绝、文件不存在包含路径、`Validate` 错误透传。

核心测试形态：

```go
Convey("Load 读取本地 YAML 并保留默认值", t, func() {
	cfg := &testConfig{Port: 8080}
	err := Load(dir, "conf", cfg)
	So(err, ShouldBeNil)
	So(cfg.Port, ShouldEqual, 8080)
	So(cfg.Name, ShouldEqual, "life-core")
})
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./pkg/hconfig -count=1`

Expected: FAIL，因为 `pkg/hconfig` 尚不存在。

- [ ] **Step 3: 实现最小加载器并迁移框架配置**

实现：

```go
type Validator interface {
	Validate() error
}

func Load(path, name string, target any) error
```

使用反射确认目标是非 nil 指针；使用独立 Viper 实例读取 `<path>/<name>.yaml`；Unmarshal 后调用 `Validate`。`internal/config.NewConfig` 先填充 Server 默认值，再调用 `hconfig.Load`。

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./pkg/hconfig ./internal/config -count=1`

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add pkg/hconfig internal/config
git diff --staged --check
git commit -m "feat: add local typed config loader"
```

### Task 2: Context Logger 与 HTTP 链路注入

**Files:**
- Create: `pkg/hlog/config.go`
- Create: `pkg/hlog/context.go`
- Create: `pkg/hlog/field.go`
- Create: `pkg/hlog/log.go`
- Create: `pkg/hlog/doc.go`
- Create: `pkg/hlog/config_test.go`
- Create: `pkg/hlog/context_test.go`
- Modify: `internal/config/config.go`
- Modify: `internal/logger/logger.go`
- Modify: `internal/middleware/request_id.go`
- Modify: `internal/middleware/logging.go`
- Modify: `internal/middleware/response.go`
- Modify: `internal/middleware/recovery.go`
- Modify: `internal/middleware/middleware_test.go`
- Modify: `hollow.go`

**Interfaces:**
- Produces: `type Logger = zap.Logger`、`type Field = zap.Field`
- Produces: `func New(Config) (*Logger, error)`
- Produces: `func SetDefault(*Logger)`、`func L() *Logger`
- Produces: `func NewContext(context.Context, *Logger) context.Context`
- Produces: `func FromContext(context.Context) *Logger`
- Produces: 常用字段函数 `String`、`Int`、`Int32`、`Int64`、`Uint64`、`Bool`、`Duration`、`Any`、`Err`

- [ ] **Step 1: 写失败测试**

用 GoConvey 覆盖默认 Logger 永不为 nil、Context 注入后返回同一实例、nil Logger 回退默认实例、非法 level/output mode 返回错误、Request ID 中间件注入带字段 Logger。

核心断言：

```go
ctx := NewContext(context.Background(), named)
So(FromContext(ctx), ShouldEqual, named)
So(FromContext(context.Background()), ShouldEqual, L())
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./pkg/hlog ./internal/middleware -count=1`

Expected: FAIL，因为 `pkg/hlog` API 尚不存在。

- [ ] **Step 3: 实现日志公共包和中间件注入**

`SetDefault` 使用 `atomic.Pointer[zap.Logger]`，初始值为 `zap.NewNop()`。Request ID 中间件执行：

```go
logger := hlog.L().With(hlog.String(RequestIDKey, requestID))
ctx := hlog.NewContext(c.Request.Context(), logger)
c.Request = c.Request.WithContext(ctx)
```

后续中间件统一通过 `hlog.FromContext(c.Request.Context())` 写日志。`hollow.NewApp` 构建 Logger 后调用 `hlog.SetDefault(log)`。

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./pkg/hlog ./internal/logger ./internal/middleware -count=1`

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add pkg/hlog internal/config internal/logger internal/middleware hollow.go
git diff --staged --check
git commit -m "feat: add context-aware structured logging"
```

### Task 3: MySQL 薄封装与 Redis 配置验证

**Files:**
- Create: `pkg/hgorm/gorm.go`
- Create: `pkg/hgorm/logger.go`
- Create: `pkg/hgorm/doc.go`
- Create: `pkg/hgorm/gorm_test.go`
- Modify: `pkg/hredis/redis.go`
- Modify: `pkg/hredis/redis_test.go`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Produces: `type hgorm.Config struct { DSN string; MaxOpenConns, MaxIdleConns int; ConnMaxLifetime, ConnMaxIdleTime, SlowThreshold time.Duration; LogLevel string }`
- Produces: `func hgorm.NewMySQL(context.Context, Config) (*gorm.DB, error)`
- Produces: `func hgorm.NewDB(context.Context, gorm.Dialector, Config) (*gorm.DB, error)`
- Produces: `func hgorm.Ping(context.Context, *gorm.DB) error`
- Consumes: `hlog.FromContext(ctx)`

- [ ] **Step 1: 写失败测试**

用 GoConvey 覆盖空 DSN、nil Dialector、默认连接池参数、Ping 错误、GORM Logger 从 Context 继承字段，以及 Redis 空地址和默认超时。

`NewDB` 测试使用 `go-sqlmock` 和 MySQL Dialector 的 `Conn` 注入，不连接真实数据库。

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./pkg/hgorm ./pkg/hredis -count=1`

Expected: FAIL，因为 `pkg/hgorm` 尚不存在或新断言未满足。

- [ ] **Step 3: 实现最小数据库封装**

`NewMySQL` 校验 DSN 后委托给 `NewDB`；`NewDB` 设置连接池并 Ping。GORM Logger 的 `Trace(ctx, ...)` 使用 `hlog.FromContext(ctx)`，按 error、slow、info 分级输出。

Redis 保持返回原生 `*redis.Client`，仅补齐可复用默认值和导出校验错误，不包装命令。

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./pkg/hgorm ./pkg/hredis -count=1`

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add pkg/hgorm pkg/hredis go.mod go.sum
git diff --staged --check
git commit -m "feat: add mysql infrastructure support"
```

### Task 4: CLI 生成配置、数据库与 GoConvey 测试

**Files:**
- Create: `cmd/hollow_cli/generator/templates/database.go.tmpl`
- Create: `cmd/hollow_cli/generator/templates/database_mysql.go.tmpl`
- Create: `cmd/hollow_cli/generator/templates/database_redis.go.tmpl`
- Create: `cmd/hollow_cli/generator/templates/config_test.go.tmpl`
- Create: `cmd/hollow_cli/generator/templates/control_health_test.go.tmpl`
- Create: `cmd/hollow_cli/generator/templates/service_health_test.go.tmpl`
- Modify: `cmd/hollow_cli/generator/templates/config.go.tmpl`
- Modify: `cmd/hollow_cli/generator/templates/conf.yaml.tmpl`
- Modify: `cmd/hollow_cli/generator/templates/main.go.tmpl`
- Modify: `cmd/hollow_cli/generator/templates/go.mod.tmpl`
- Modify: `cmd/hollow_cli/generator/templates/control_health.go.tmpl`
- Modify: `cmd/hollow_cli/generator/render.go`
- Modify: `cmd/hollow_cli/generator/project_test.go`
- Modify: `cmd/hollow_cli/generator/integration_test.go`
- Modify: `cmd/hollow_cli/generator/templates/README.md.tmpl`

**Interfaces:**
- Produces: `config.Startup()`、`config.Get()`
- Produces: `database.InitMySQL()`、`database.InitRedis()`、`database.Shutdown(ctx)`
- Produces: `database.MySQL()`、`database.Redis()`
- Consumes: `hconfig.Load`、`hgorm.NewMySQL`、`hredis.NewClient`

- [ ] **Step 1: 扩展 CLI 失败测试**

Golden Tree 必须包含：

```text
config/config_test.go
control/health_test.go
database/database.go
database/mysql.go
database/redis.go
service/health_test.go
```

模板断言必须覆盖 `enabled`、`config.Startup` 排在数据库之前、Shutdown、GoConvey 依赖和 `hlog.FromContext(ctx)`。

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./cmd/hollow_cli/generator -count=1`

Expected: FAIL，缺少新增模板文件。

- [ ] **Step 3: 实现模板**

生成项目启动顺序：

```go
app.Startup(
	config.Startup,
	database.InitMySQL,
	database.InitRedis,
)
app.Shutdown(database.Shutdown)
```

`conf.yaml` 中两个数据库默认 `enabled: false`。Control Health 使用 Context Logger；三个生成测试使用 GoConvey。

- [ ] **Step 4: 运行 CLI 单测和生成项目集成测试**

Run: `go test ./cmd/hollow_cli/generator -count=1`

Expected: PASS，包括临时项目 `make proto → go mod tidy → go test → go build`。

- [ ] **Step 5: 提交**

```bash
git add cmd/hollow_cli/generator
git diff --staged --check
git commit -m "feat: generate application infrastructure"
```

### Task 5: 同步 life-core 参考项目

**Files:**
- Create: `/Users/nikki/go/src/life-core/database/database.go`
- Create: `/Users/nikki/go/src/life-core/database/mysql.go`
- Create: `/Users/nikki/go/src/life-core/database/redis.go`
- Create: `/Users/nikki/go/src/life-core/config/config_test.go`
- Create: `/Users/nikki/go/src/life-core/control/health_test.go`
- Create: `/Users/nikki/go/src/life-core/service/health_test.go`
- Modify: `/Users/nikki/go/src/life-core/config/config.go`
- Modify: `/Users/nikki/go/src/life-core/conf.yaml`
- Modify: `/Users/nikki/go/src/life-core/main.go`
- Modify: `/Users/nikki/go/src/life-core/control/health.go`
- Modify: `/Users/nikki/go/src/life-core/go.mod`
- Modify: `/Users/nikki/go/src/life-core/go.sum`
- Modify: `/Users/nikki/go/src/life-core/README.md`

**Interfaces:**
- Consumes: Hollow 本轮新增公共 API
- Produces: 可运行的 `GET /v1/health` 参考项目

- [ ] **Step 1: 先复制 CLI 目标断言到 life-core 测试并确认失败**

Run: `go test ./... -count=1`

Expected: FAIL，因为新配置和数据库基础设施尚未同步。

- [ ] **Step 2: 按模板同步文件**

保持 MySQL、Redis 默认关闭；健康 Control 通过 `hlog.FromContext(ctx)` 写结构化日志。

- [ ] **Step 3: 生成 Proto 并整理依赖**

Run: `make proto && go mod tidy`

Expected: 命令退出码为 0。

- [ ] **Step 4: 验证 life-core**

Run: `go test ./... -count=1 && go build ./...`

Expected: PASS。

`life-core` 当前不是 Git 仓库，因此不执行提交。

### Task 6: 文档、兼容性与最终验证

**Files:**
- Modify: `README.md`
- Modify: `claude.md`
- Modify: `cmd/hollow_cli/generator/templates/README.md.tmpl`

**Interfaces:**
- Documents: `hconfig`、`hlog`、MySQL/Redis 生命周期、分层和可选 Cache

- [ ] **Step 1: 更新使用文档**

文档给出以下日志示例：

```go
logger := hlog.FromContext(ctx).Named("CreatorService.GetHomepage").
	With(hlog.String("open_id", openID))
```

同时说明配置只在启动时读取、数据库 `enabled` 行为和 Cache 适用边界。

- [ ] **Step 2: 执行格式和静态检查**

Run: `gofmt -w <本轮修改的 Go 文件>`

Run: `go vet ./...`

Expected: PASS。

- [ ] **Step 3: 执行 Hollow 全量测试**

Run: `go test ./... -count=1`

Expected: PASS。

- [ ] **Step 4: 执行 Race 测试**

Run: `go test -race ./... -count=1`

Expected: PASS。

- [ ] **Step 5: 启动 life-core 做 HTTP 验收**

MySQL/Redis 禁用时启动服务并验证：

```text
GET /v1/health  → HTTP 200，code=0，data.status=ok
GET /-/health   → HTTP 404，code=1200
```

- [ ] **Step 6: 提交最终文档**

```bash
git add README.md claude.md cmd/hollow_cli/generator/templates/README.md.tmpl
git diff --staged --check
git commit -m "docs: document application foundation"
```
