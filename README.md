# hollow 轻量级web框架

基于go-gin框架封装轻量级的web框架，提供了开箱即用的功能，主要是提升编码能力和沉淀go成熟的库
# 项目结构

```
hollow/
├── cmd/                          # 命令行工具（脚手架）
│   └── hollow-cli/               # IDL代码生成器
│       ├── main.go               # CLI入口（Cobra）
│       └── generator/            # 代码生成逻辑
│           ├── proto.go          # Protobuf解析生成
│           └── thrift.go         # Thrift解析生成（预留）
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
    ├── api/                      # 业务API
    │   ├── user/                 # 用户服务
    │   │   ├── user.proto        # Protobuf定义
    │   │   ├── user.pb.go        # 自动生成代码
    │   │   └── handler.go        # 业务逻辑
    │   └── config.yaml           # 配置文件示例
    └── main.go                   # 服务启动入口
```
# 核心功能
## 1. 框架核心 (hollow.go)
- App 结构体 ：框架的核心，管理整个应用生命周期
- 中间件管理 ：支持动态添加/移除中间件，自动去重
- 优雅启停 ：通过信号处理实现优雅关闭
- 依赖注入 ：支持用户自定义配置和中间件
## 2. 配置管理 (config.go)
- 基于 Viper 实现，支持 YAML 配置文件
- 支持日志、数据库、Redis 等配置
- 配置结构化管理
## 3. 日志系统 (logger.go)
- 基于 Zap 高性能日志库
- 支持 Console 和 File 两种输出模式
- 自动日志轮转
- 支持 Debug/Info/Warn/Error 多级别
## 4. 中间件系统 (middleware/)
采用接口化设计，每个中间件实现 Middleware 接口：
- Handle ：处理请求，返回响应
- Name ：中间件名称，用于日志记录

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
- 常用时间格式常量 hredlock - 分布式锁
- 基于 Redis 的分布式锁实现
- 适用于秒杀等高并发场景 hexcel - Excel 处理
- Excel 读取和解析
- Excel 转 SQL 工具 hes - Elasticsearch 客户端
- ES 连接和操作封装

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