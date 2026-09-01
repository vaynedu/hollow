# Hollow Control-Service-DAO-Model 脚手架设计

## 背景

当前 Hollow 脚手架生成的项目由 `proto/*_myhttp.pb.go` 完成 HTTP 路由、参数绑定和统一响应适配，并让一级目录 `service/` 直接实现 Proto Service 接口。这使接口控制职责与核心业务逻辑混在同一层，不符合现有 Sparrow 项目常用的分层方式。

本次调整只改变新项目的业务目录和装配关系，不改变 Hollow 核心生命周期、中间件和统一响应协议，也不改变 `protoc-gen-myhttp` 的生成职责。

## 目标

新生成项目采用以下一级目录：

```text
project/
├── control/       # 接口控制层，手写
├── service/       # 核心业务逻辑层，手写
├── dao/           # 数据访问层，手写
├── model/         # 数据库模型层，手写
├── proto/         # Proto 定义及自动生成代码
├── router/        # 路由装配
├── config/
├── main.go
├── Makefile
└── conf.yaml
```

调用链固定为：

```text
HTTP 请求
→ proto/*_myhttp.pb.go
→ control
→ service
→ dao
→ model
```

## 职责边界

### 自动生成的 HTTP Adapter

`protoc-gen-myhttp` 继续生成 `proto/*_myhttp.pb.go`，负责：

- 将 Proto HTTP 注解转换为 Gin 路由；
- 绑定查询参数、JSON 请求体和路径参数；
- 调用 Proto Service 接口；
- 将数据或错误交给 Hollow 统一响应中间件。

该文件是生成代码，禁止手工修改。项目不增加独立 `handler/` 目录，因为 Handler 仅承担机械适配职责。

### control

`control/` 实现 Proto Service 接口，是用户编写的接口控制层，负责：

- 接收已完成绑定的 Proto Request；
- 做接口级参数和权限校验；
- 编排一个或多个 Service；
- 将领域结果转换成 Proto Response；
- 返回 Hollow 可识别的业务错误。

Control 方法使用 `context.Context`，不直接依赖 `gin.Context`：

```go
func (Control) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error)
```

脚手架初始化时只生成 `control/control.go`，其中包含单例和 `proto.Unimplemented<Service>Service` 嵌入。具体 RPC 方法由开发者按业务创建，不由 `make proto` 写入或覆盖。

### service

`service/` 实现可复用的核心业务逻辑，不处理 Gin 请求绑定，也不负责 HTTP 响应格式。初始化时生成仅含包声明和说明的 `service/service.go`。

### dao

`dao/` 封装数据库查询和持久化操作。初始化时只创建空骨架，不预设 MySQL、GORM、Redis 或全局数据库实例，避免通用脚手架与具体基础设施绑定。

### model

`model/` 保存数据库持久化模型。初始化时只创建空骨架，不预设 GORM 基类或自动迁移行为。

### router

`router/` 保留显式装配职责，但由原来的 `service.Get()` 改为 `control.Get()`：

```go
proto.Register<Service>GinRouter(app.Engine, control.Get())
```

## 生成和覆盖规则

- `hollow-cli init` 一次性生成完整项目目录和手写骨架；
- 目标项目目录必须不存在，现有的原子发布行为保持不变；
- `make proto` 只运行 `protoc-gen-go` 和 `protoc-gen-myhttp`；
- `make proto` 只更新 `proto/*.pb.go` 与 `proto/*_myhttp.pb.go`；
- `control/`、`service/`、`dao/`、`model/` 永远不由 `make proto` 修改；
- 不新增用于逐个 RPC 生成手写 Control 方法的 CLI 命令。

## 对现有项目的影响

Hollow 核心 API 和 `protoc-gen-myhttp` 输出接口不变，因此已有项目可继续编译。新模板生效后：

- 新项目直接使用新分层；
- 已有 `life-core` 需要将 Proto 接口实现从 `service/` 移至 `control/`；
- `router` 改为装配 `control.Get()`；
- 新增空的 `dao/` 与 `model/` 骨架；
- 原 `service/` 保留，转换为纯业务逻辑层。

## 错误处理

- HTTP 参数绑定错误继续由生成 Adapter 包装为 `hecode.ErrInvalidParam`；
- Control 和 Service 返回的错误继续交给 Hollow 统一响应中间件；
- DAO 返回原始数据访问错误，由 Service 或 Control 按业务语义转换；
- 本次不新增错误码体系或 Gin 专用 Control API。

## 验证标准

1. CLI 模板测试确认生成 `control/service/dao/model` 四个一级目录；
2. 生成的 Router 引用 `control.Get()`，不再引用 `service.Get()`；
3. 生成项目执行 `make proto` 后，手写目录内容保持不变；
4. 新项目可依次通过 `make proto`、`make deps`、`make test`、`make build`；
5. Hollow 根模块全量测试和 Race 测试通过；
6. 将现有 `life-core` 迁移到新结构后测试和构建通过。

## 非目标

- 不修改 Hollow 核心生命周期和中间件；
- 不修改统一响应格式；
- 不引入 IoC、Repository 接口、代码生成 DAO 或 GORM 模型；
- 不生成独立 `handler/` 或 `server/` 目录；
- 不让 Control 依赖 `gin.Context`；
- 不在本次实现具体业务逻辑或数据库连接。
