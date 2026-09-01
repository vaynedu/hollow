# hollow 轻量级web框架

基于go-gin框架封装轻量级的web框架，提供了开箱即用的功能，主要是提升编码能力和沉淀go成熟的库

# 快速开始（当前未发布阶段）

环境要求：Go 1.25+、Protocol Buffers 编译器，以及以下代码生成插件：

```bash
git clone https://github.com/googleapis/googleapis.git /absolute/path/to/googleapis
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.1
(cd /path/to/protoc-gen-myhttp && go install .)
```

`protoc-gen-myhttp` 必须从已合入本次 HTTP adapter 修复的源码目录安装，不要使用尚未包含修复的 latest 版本。
`google/api/annotations.proto` 来自 googleapis 仓库（也可使用已有安装）；`PROTO_INCLUDE` 必须指向 googleapis 根目录，且 `/absolute/path/to/googleapis/google/api/annotations.proto` 必须存在。

当前发布的 `v1.0.0` 尚未包含最新的生命周期、错误码、配置、日志和数据库 API，本地开发必须显式指定当前 Hollow 源码：

```bash
hollow-cli init <project> --hollow-path /path/to/hollow
cd <project>
make proto PROTO_INCLUDE=/absolute/path/to/googleapis
```

只有发布包含新 API 的 Hollow 版本，并将 CLI 的 `DefaultHollowVersion` 更新后，才可使用 `hollow-cli init <project>` 省略 `--hollow-path`。现有 `v1.0.0` 标签不得移动或覆盖。

# 项目结构

```
hollow/
├── cmd/                          # 命令行工具（脚手架）
│   └── hollow_cli/               # 项目脚手架工具
│       ├── main.go               # CLI入口
│       ├── command.go            # init 命令定义
│       └── generator/            # 项目骨架生成逻辑
│           ├── project.go        # 项目配置与目录生成
│           ├── render.go         # 嵌入模板渲染
│           └── templates/        # 项目骨架模板
├── internal/                     # 框架内部配置、日志和 HTTP 中间件
├── pkg/                          # 对业务项目公开的基础能力
│   ├── hconfig/                  # 本地 YAML 类型化加载
│   ├── hlog/                     # Context 结构化日志
│   ├── hgorm/                    # GORM MySQL 与连接池
│   ├── hredis/                   # Redis 客户端
│   └── hecode/                   # 统一错误码与响应
├── hollow.go                     # 框架入口
└── example/                      # 使用示例
    ├── config/config.go          # 类型化业务配置
    ├── database/                 # MySQL、Redis 生命周期
    ├── proto/                    # Protobuf 定义和生成代码
    ├── router/router.go          # 生成路由注册
    ├── service/service.go        # Service 实现骨架
    ├── Makefile                  # 代码生成与依赖整理
    ├── conf.yaml                 # 配置文件示例
    └── main.go                   # 服务启动入口
```

生成并验证示例：

```bash
make -C example proto PROTO_INCLUDE=/absolute/path/to/googleapis
make -C example deps
(cd example && go test ./... && go build ./...)
```

`make proto` 只生成 `*.pb.go` 和 `*_myhttp.pb.go`；`make deps` 显式整理 Go module 依赖。

`hollow-cli init` 生成的业务项目采用一级目录分层：

```text
proto/*_myhttp.pb.go（自动生成 HTTP Adapter）
→ control（接口校验、业务编排和响应组装）
→ service（核心业务逻辑）
→ dao（数据访问）
→ model（数据库模型）

service → cache（可选）
```

项目不生成独立 `handler/` 目录。HTTP Adapter 保留在 `proto/*_myhttp.pb.go`，`make proto` 不会修改 `control/`、`service/`、`dao/`、`model/` 中的手写代码。`cache/` 仅在读多写少且允许短暂不一致时按需创建。

生成的 `model/user.go` 是数据库模型示例，展示显式 GORM/JSON 标签、表名常量和状态枚举；业务项目可按真实表结构修改或替换。

新项目默认通过 Proto 提供 `GET /v1/health`，由 `control.Health` 调用 `service.Health`。Hollow 核心不注册服务级健康路由。

# 核心功能
## 1. 框架核心 (hollow.go)
- App 结构体 ：框架的核心，管理整个应用生命周期
- 中间件管理 ：通过 `app.Use(...gin.HandlerFunc)` 注册 Gin 原生中间件
- 优雅启停 ：通过信号处理实现优雅关闭
- 依赖注入 ：支持用户自定义配置和中间件

```go
app, err := hollow.NewApp(hollow.AppOption{ConfigPath: ".", ConfigName: "conf"})
if err != nil {
	return err
}

app.Startup(
	config.Startup,
	database.InitMySQL,
	database.InitRedis,
)
app.Shutdown(database.Shutdown)
router.Register(app)
return app.Run()
```

`Startup(...)` 按注册顺序执行启动 Hook，`Shutdown(...)` 按注册的逆序执行关闭 Hook。`Run()` 启动 HTTP Server，并在收到退出信号后完成优雅关闭。

## 2. 配置管理 (`pkg/hconfig`)
- 启动时读取本地 `conf.yaml`
- 支持类型化反序列化、调用方默认值和 `Validate()` 校验
- 本期不包含热更新、远程配置或环境变量覆盖

## 3. 日志系统 (`pkg/hlog`)
- 基于 Zap 高性能日志库
- 支持 Console 和 File 两种输出模式
- 支持日志文件轮转、保留份数和压缩
- 支持 Debug/Info/Warn/Error 多级别
- Request ID 自动注入 Context，可在任意业务层直接使用：

```go
logger := hlog.FromContext(ctx).Named("CreatorService.GetHomepage").
	With(hlog.String("open_id", openID))
logger.Info("homepage loaded")
```

## 4. 数据库 (`pkg/hgorm`、`pkg/hredis`)

- `hgorm.NewMySQL`：创建 GORM MySQL 客户端、配置连接池、Ping，并把 SQL 日志接入 Context Logger。
- `hredis.NewClient`：创建原生 go-redis v9 客户端并 Ping。
- CLI 默认生成 `database/mysql.go`、`database/redis.go`；通过 `conf.yaml` 的 `enabled` 开关启用。

## 5. 中间件系统 (middleware/)
采用 Gin 原生中间件模型；框架内置中间件和业务自定义中间件均为 `gin.HandlerFunc`，可通过 `app.Use(...)` 注册：

- RequestID ：请求追踪 ID 生成
- Logging ：请求日志记录（方法、路径、耗时、状态码等）
- Recovery ：Panic 恢复，防止服务崩溃
- Response ：统一响应格式处理
- Metrics ：性能指标收集（可扩展）

## 6. 工具包 hcond - 条件构造器
- 支持构建复杂的 SQL WHERE 条件
- 支持逻辑运算符（AND/OR）
- 支持比较运算符（=, !=, >, <, >=, <=, IN）
- 自动生成 SQL 和参数绑定 hidgenerator - ID 生成器
- UUID 生成
- 接口化设计，易于扩展其他 ID 生成策略 hresty - HTTP 客户端
- 基于 Resty 封装
- 详细的请求追踪（DNS 查询、TCP 连接、TLS 握手等）
- 连接池管理
- 性能优化配置 hlark - 飞书集成
- Webhook 消息发送
- HMAC 签名验证
- 支持文本消息推送 htime - 时间处理
- 时间戳转换（秒级、毫秒级）
- 时间格式化
- 常用时间格式常量

hredlock - 分布式锁
- 基于 go-redsync/v4,Lock/Unlock 二件套
- 接受外部 *redis.Client,锁 expiry 强制显式传入(防死锁)

hcsv - CSV 文件读取
- 一次性加载 CSV 到内存,适合配置/小数据

hes - Elasticsearch 客户端
- 基于 elasticsearch/v8,Config + NewClient(自带 Info 探活) + Ping 薄封装
- 拿到原生 *elasticsearch.Client 后直接使用 v8 esapi

# 技术栈
- Web 框架 ：Gin
- 配置管理 ：Viper
- 日志 ：Zap + Lumberjack
- 数据库 ：GORM
- 缓存 ：Redis (go-redis)
- HTTP 客户端 ：Resty
- 搜索 ：Elasticsearch
- 分布式锁 ：Redsync
- CLI 工具 ：Cobra
