# hollow 轻量级web框架

基于go-gin框架封装轻量级的web框架，提供了开箱即用的功能，主要是提升编码能力和沉淀go成熟的库

# 快速开始（当前未发布阶段）

环境要求：Go 1.25+、Protocol Buffers 编译器，以及以下代码生成插件：

```bash
git clone https://github.com/googleapis/googleapis.git /absolute/path/to/googleapis
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
(cd /path/to/protoc-gen-myhttp && go install .)
```

`protoc-gen-myhttp` 必须从已合入本次 HTTP adapter 修复的源码目录安装，不要使用尚未包含修复的 latest 版本。
`google/api/annotations.proto` 来自 googleapis 仓库（也可使用已有安装）；`PROTO_INCLUDE` 必须指向 googleapis 根目录，且 `/absolute/path/to/googleapis/google/api/annotations.proto` 必须存在。

当前发布的 `v1.0.0` 尚未同时包含 `Startup/Shutdown/Run` 与 `pkg/hecode`，本地开发必须显式指定当前 Hollow 源码：

```bash
hollow-cli init <project> --hollow-path /path/to/hollow
cd <project>
make proto PROTO_INCLUDE=/absolute/path/to/googleapis
```

只有发布包含新 API 的新 Hollow 版本（即包含 `Startup/Shutdown/Run` 与 `pkg/hecode`），并将 CLI 的 `DefaultHollowVersion` 更新到该新版本后，才可使用 `hollow-cli init <project>` 省略 `--hollow-path`。现有 `v1.0.0` 标签不得移动或覆盖。

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
├── internal/                     # 框架核心实现（不对外暴露）
│   ├── config/                   # 配置管理（本地+远程热加载）
│   │   ├── loader.go             # 配置加载器
│   │   └── watcher.go            # 热加载监听
│   ├── logger/                   # 日志模块（Zap封装）
│   │   └── logger.go             # 日志初始化
│   ├── metrics/                  # 打点上报（Prometheus）
│   │   └── metrics.go            # 指标收集
│   ├── middleware/               # 核心中间件
│   │   ├── response.go           # 统一响应
│   │   ├── recovery.go           # 错误恢复
│   │   └── logging.go            # 日志记录
│   ├── router/                   # 路由注册
│   │   └── router.go             # HTTP路由绑定
│   └── grpc/                     # gRPC扩展（预留）
│       └── server.go             # gRPC服务器
├── pkg/                          # 公共工具库（对外暴露）
│   ├── conv/                     # 数字转换工具
│   │   └── conv.go               # 类型转换
│   ├── pool/                     # 协程池
│   │   └── worker_pool.go        # 任务池实现
│   ├── retry/                    # 重试机制
│   │   └── retry.go              # 带退避的重试
│   └── once/                     # 仅运行一次
│       └── once.go               # sync.Once封装
├── hollow.go                     # 框架入口
└── example/                      # 使用示例
    ├── config/config.go          # 业务生命周期 Hook
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

app.Startup(initDependencies)
app.Shutdown(closeDependencies)
router.Register(app)
return app.Run()
```

`Startup(...)` 按注册顺序执行启动 Hook，`Shutdown(...)` 按注册的逆序执行关闭 Hook。`Run()` 启动 HTTP Server，并在收到退出信号后完成优雅关闭。

## 2. 配置管理 (config.go)
- 基于 Viper 实现，支持 YAML 配置文件
- 支持 HTTP Server 和日志配置
- 配置结构化管理
## 3. 日志系统 (logger.go)
- 基于 Zap 高性能日志库
- 支持 Console 和 File 两种输出模式
- 自动日志轮转
- 支持 Debug/Info/Warn/Error 多级别
## 4. 中间件系统 (middleware/)
采用 Gin 原生中间件模型；框架内置中间件和业务自定义中间件均为 `gin.HandlerFunc`，可通过 `app.Use(...)` 注册：

- RequestID ：请求追踪 ID 生成
- Logging ：请求日志记录（方法、路径、耗时、状态码等）
- Recovery ：Panic 恢复，防止服务崩溃
- Response ：统一响应格式处理
- Metrics ：性能指标收集（可扩展）

## 5. 工具包 hcond - 条件构造器
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
