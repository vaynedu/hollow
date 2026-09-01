# Hollow 应用基础能力设计

## 1. 目标

Hollow 提供通用 Web 服务所需的配置、日志、MySQL、Redis 和请求上下文能力；`hollow-cli init` 生成可直接运行、可测试的项目骨架。业务项目只需要围绕 `control → service → dao → model` 开发业务逻辑，按需增加 `cache`。

本设计同时升级 `/Users/nikki/go/src/life-core`，将它作为脚手架的真实验收项目。

## 2. 设计原则

- 保持轻量，不引入 IoC、服务发现、远程配置、gRPC 或 Worker。
- 配置仅从本地 `conf.yaml` 读取一次；本期不支持热更新、远程配置和环境变量覆盖。
- 框架负责通用基础设施，业务项目负责业务配置和业务对象。
- MySQL、Redis 文件默认生成，通过配置中的 `enabled` 决定是否初始化。
- `cache` 不是固定分层，仅在读多写少且允许短暂不一致时由业务项目添加。
- 所有新增单元测试使用 GoConvey 的 `Convey`、`So` 风格。

## 3. 分层与依赖方向

```text
Proto HTTP Adapter
        ↓
control       接收请求、调用 service、返回结果
        ↓
service       参数校验、业务编排、错误转换、降级策略
        ↓
dao           封装数据库读写
        ↓
model         定义数据库模型

service → cache（可选）
```

约束：

- Control 不直接访问数据库或缓存。
- Service 不保存请求级可变状态；请求参数在一次方法调用内传递。
- DAO 返回数据库错误，Service 将其转换为业务错误码。
- Model 不依赖 Control、Service 或 DAO。
- Cache 由 Service 通过显式依赖使用，不成为所有项目的必选目录。

`czg-core/GetCreatorHomepage` 中值得保留的经验是上下文日志、显式依赖、主流程与可降级依赖的区分，以及 DAO/Model 边界。其 `InitAndCheckParam` 写入 Service 字段的两阶段调用不复制到 Hollow，因为这会让 Service 带请求状态并增加并发复用风险。

## 4. 配置设计

### 4.1 Hollow 公共加载器

新增 `pkg/hconfig`，提供：

```go
type Validator interface {
	Validate() error
}

func Load(path, name string, target any) error
```

行为：

1. 只读取 `<path>/<name>.yaml`。
2. `target` 必须是非 nil 指针。
3. 调用方可以先构造带默认值的对象，再由 YAML 覆盖。
4. 如果对象实现 `Validator`，反序列化后自动校验。
5. 错误包含配置文件路径和具体阶段，便于定位。

Hollow 自身的 Server/Log 配置改用同一加载器，保留现有 `ConfigPath`、`ConfigName` API，避免破坏已有项目。

### 4.2 生成项目配置

CLI 生成的 `config/config.go` 定义类型化配置和只读全局快照：

```go
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
}

func Startup() error
func Get() *Config
```

`Startup` 读取本地 `conf.yaml`。因为本期只在启动时加载，所以不引入原子热更新和文件监听。

MySQL、Redis 配置均包含 `enabled`。关闭时跳过连接，开启但缺少必填地址时启动失败并返回清晰错误。

## 5. 日志设计

新增公共包 `pkg/hlog`，封装 zap 并提供 Sparrow 风格 API：

```go
hlog.FromContext(ctx).
	Named("RestaurantService.List").
	With(hlog.String("open_id", openID)).
	Info("list restaurant success")
```

公共能力包括：

- `L()`：获得进程默认 Logger。
- `SetDefault(logger)`：由 Hollow App 初始化默认 Logger。
- `NewContext(ctx, logger)`：显式注入 Logger。
- `FromContext(ctx)`：读取 Logger，未注入时安全回退到默认 Logger。
- 常用 zap 类型和字段构造器别名，如 `String`、`Int64`、`Duration`、`Any`、`Err`。

Request ID 中间件负责创建带 `request_id` 字段的 Logger 并写入 `http.Request.Context()`。Control、Service、DAO 使用同一个 Context 后，无需重复拼接 Request ID。HTTP 日志、Recovery 和统一错误日志也改为使用 Context Logger。

日志配置继续支持控制台或文件、级别、文件轮转大小、保留天数，并补充最大备份数和压缩开关。非法日志级别或输出模式应在启动时失败，不静默回退。

## 6. MySQL 与 Redis

### 6.1 Hollow

新增 `pkg/hgorm`：

- `Config`：DSN、连接池、连接生命周期、慢 SQL 阈值、日志级别。
- `NewMySQL(ctx, cfg)`：创建 GORM MySQL 连接、设置连接池、执行 Ping。
- `Ping(ctx, db)`：健康检查。
- GORM Logger 将 SQL 日志写入 `hlog.FromContext(ctx)`，自动继承 Request ID。

保留并补强现有 `pkg/hredis`：

- 继续返回原生 `*redis.Client`，不包装 Redis 命令。
- 使用配置默认值并在创建时 Ping。
- 明确配置校验和关闭行为。

### 6.2 CLI 生成项目

新增：

```text
database/database.go
database/mysql.go
database/redis.go
```

公开接口：

```go
func Startup() error
func Shutdown(context.Context) error
func MySQL() *gorm.DB
func Redis() *redis.Client
```

初始化顺序为 MySQL → Redis；任一步失败时关闭已创建资源。关闭顺序相反。禁用的组件保持 nil，业务代码在启用后才能使用对应 Getter。

`main.go` 的生命周期为：

```go
app.Startup(
	config.Startup,
	database.Startup,
)
app.Shutdown(database.Shutdown)
```

## 7. CLI 与 life-core

CLI 的固定目录增加 `database/`，但不增加组件选择参数。默认 `conf.yaml` 同时展示 Server、Log、MySQL、Redis 配置；MySQL 和 Redis 默认关闭，确保新项目无需外部服务即可运行 `/v1/health`。

CLI 同时生成以下 GoConvey 测试：

- `config/config_test.go`：默认配置、YAML 覆盖、非法启用配置。
- `service/health_test.go`：健康业务返回 `ok`。
- `control/health_test.go`：Control 正确调用 Service 并组装响应。

`life-core` 同步上述文件、配置和日志调用，继续保留 `GET /v1/health` 作为端到端验证接口。

## 8. 测试与验收

Hollow：

- `pkg/hconfig` 覆盖成功加载、默认值保留、目标非法、文件不存在、校验失败。
- `pkg/hlog` 覆盖默认 Logger、Context 注入、Request ID 传播。
- `pkg/hgorm` 使用 mock dialector/SQL mock 或最小可控驱动验证配置，不依赖真实 MySQL。
- `pkg/hredis` 单测不依赖真实 Redis。
- CLI Golden Tree、模板内容和集成测试覆盖新增文件。

集成验收：

1. Hollow 执行 `go test ./... -count=1`。
2. Hollow 执行 `go test -race ./... -count=1`。
3. CLI 生成临时项目，执行 `make proto`、`go mod tidy`、`go test ./...`、`go build ./...`。
4. `life-core` 执行 `make proto`、`go test ./...`、`go build ./...`。
5. 在 MySQL/Redis 禁用时启动 `life-core`，验证 `/v1/health` 返回 HTTP 200、业务码 0；未知路由返回 HTTP 404、业务码 1200。

## 9. 非目标

- 不实现配置热更新、远程配置或环境变量覆盖。
- 不引入 IoC 容器。
- 不生成具体业务 DAO、Model 或缓存实现。
- 不要求本机存在 MySQL、Redis 才能运行健康检查和单元测试。
- 不修改 `protoc-gen-myhttp` 的职责和生成位置。
