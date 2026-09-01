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
app, err := hollow.NewApp(hollow.AppOption{ConfigPath: ".", ConfigName: "conf"})
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

## 示例 API
- `POST /v1/users`：创建用户
- `GET /v1/users?id=...`：获取用户
