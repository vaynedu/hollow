# Hollow 框架技术文档

## 项目概述
Hollow 是基于 Go 的轻量级 Web 框架，中间件架构，提供代码生成能力，快速构建 RESTful API 服务。

## 核心技术
- Go 1.23+
- Gin v1.10.0
- Protocol Buffers 3.0+

## 主要功能
- **项目初始化**：`hollow-cli init project_name` 自动创建完整结构
- **代码生成**：`hollow-cli proto proto/service.proto` 生成 Handler、Service、Router
- **标准响应**：统一 API 响应格式 `{code, msg, request_id, data}`
- **中间件**：响应格式化、异常恢复、请求日志

## 快速开始
1. 安装：`apt install protobuf-compiler`，`go install protoc-gen-go protoc-gen-myhttp`
2. 创建：`hollow-cli init myapp && cd myapp`
3. 定义：创建 `proto/user.proto` 定义服务和方法
4. 生成：`make proto` 生成代码
5. 运行：`make run` 启动服务

## 应用生命周期

```go
app, err := hollow.NewApp(hollow.AppOption{ConfigPath: ".", ConfigName: "conf"})
if err != nil {
	return err
}

app.Startup(initDependencies)
app.Shutdown(closeDependencies)
router.RegisterRoutes(app)
return app.Run()
```

- `Startup(...)`：注册应用启动 Hook，按注册顺序执行。
- `Shutdown(...)`：注册应用关闭 Hook，按注册的逆序执行。
- `router.RegisterRoutes(app)`：在 HTTP Server 启动前完成路由注册。
- `Run()`：启动 HTTP Server，并在收到退出信号后完成优雅关闭。

## 示例 API
- `POST /v1/users`：创建用户
- `GET /v1/users`：查询列表
- `GET /v1/users/{id}`：获取详情
- `PUT /v1/users/{id}`：更新用户
- `DELETE /v1/users/{id}`：删除用户
