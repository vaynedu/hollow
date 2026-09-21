# Hollow 框架技术文档

## 项目概述
Hollow 是基于 Go 的轻量级 Web 框架，中间件架构，提供项目脚手架能力，快速构建 RESTful API 服务。

## 核心技术
- Go 1.25+
- Gin v1.10.0
- Protocol Buffers 3.0+

## 主要功能
- **项目初始化**：当前使用 `hollow-cli init <project> --hollow-path /path/to/hollow` 创建完整结构
- **标准响应**：统一 API 响应格式 `{code, msg, request_id, data}`
- **中间件**：响应格式化、异常恢复、请求日志
- **配置**：`pkg/hconfig` 加载本地 YAML，支持环境变量覆盖和类型校验
- **日志**：`pkg/hlog` 从 `context.Context` 获取自动携带 Request ID 的结构化 Logger
- **数据库**：`pkg/hgorm` 和 `pkg/hredis` 提供 MySQL、Redis 薄封装
- **定时任务**：`pkg/hscheduler` 提供 Cron、超时、防重入、启动补跑和优雅关闭

## 快速开始
1. 安装 Protocol Buffers 编译器和代码生成插件：

   ```bash
   git clone https://github.com/googleapis/googleapis.git /absolute/path/to/googleapis
   go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.1
   (cd /path/to/protoc-gen-myhttp && go install .)
   ```

   `protoc-gen-myhttp` 必须从已合入本次 HTTP adapter 修复的源码目录安装，不要使用尚未包含修复的 latest 版本。
   `google/api/annotations.proto` 来自 googleapis 仓库（也可使用已有安装）；`PROTO_INCLUDE` 必须指向 googleapis 根目录，且 `/absolute/path/to/googleapis/google/api/annotations.proto` 必须存在。

2. 创建：`hollow-cli init myapp --hollow-path /path/to/hollow && cd myapp`
3. 定义：创建 `proto/user.proto` 定义服务和方法
4. 生成：`make proto PROTO_INCLUDE=/absolute/path/to/googleapis` 生成代码
5. 依赖：`make deps` 整理 Go module 依赖
6. 运行：`make run` 启动服务

当前发布的 `v1.0.0` 尚未同时包含 `Startup/Shutdown/Run` 与 `pkg/hecode`，因此未发布阶段必须传 `--hollow-path`。只有发布包含新 API 的新 Hollow 版本（即包含 `Startup/Shutdown/Run` 与 `pkg/hecode`），并将 CLI 的 `DefaultHollowVersion` 更新到该新版本后，才可使用无该参数的默认流程。现有 `v1.0.0` 标签不得移动或覆盖。

## 应用生命周期

```go
app, err := hollow.NewApp(hollow.AppOption{
	ConfigPath: ".",
	ConfigName: "conf",
	EnvPrefix:  config.EnvPrefix,
})
if err != nil {
	return err
}

app.Startup(initDependencies)
app.Shutdown(closeDependencies)
router.Register(app)
return app.Run()
```

- `Startup(...)`：注册应用启动 Hook，按注册顺序执行。
- `Shutdown(...)`：注册应用关闭 Hook，按注册的逆序执行。
- `router.Register(app)`：在 HTTP Server 启动前完成生成路由注册。
- `Run()`：启动 HTTP Server，并在收到退出信号后完成优雅关闭。

CLI 生成项目默认使用：

```go
app.Startup(
	config.Startup,
	database.InitMySQL,
	database.InitRedis,
)
app.Shutdown(database.Shutdown)
```

MySQL、Redis 通过 `conf.yaml` 的 `enabled` 开关控制，默认关闭。配置只在进程启动时读取一次，
CLI 生成的项目允许使用项目名前缀的环境变量覆盖 YAML，例如 `MYAPP_DATABASE_MYSQL_DSN`。

简单定时任务可直接注册闭包，并使用 Hollow 生命周期启动和关闭：

```go
scheduler.RegisterFunc("cleanup", "0 3 * * *", time.Minute, false, cleanup)
app.Startup(scheduler.Startup)
app.Shutdown(scheduler.Shutdown)
```

业务日志直接从 Context 获取：

```go
logger := hlog.FromContext(ctx).Named("RestaurantService.List")
logger.Info("restaurant list loaded", hlog.String("open_id", openID))
```

推荐依赖方向为 `control → service → dao → model`；`cache` 按业务场景添加，并且只能由 Service 使用。Service 不保存请求级可变状态。

## 示例 API
- `POST /v1/users`：创建用户
- `GET /v1/users?id=...`：获取用户
