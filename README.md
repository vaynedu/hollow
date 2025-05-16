# hollow 轻量级web框架

基于go-gin框架封装轻量级的web框架，主要是提升编码能力和沉淀go成熟的库

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