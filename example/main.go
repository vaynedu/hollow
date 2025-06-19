package main

import (
	"github.com/vaynedu/hollow/example/handler"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow"
	"github.com/vaynedu/hollow/internal/config"
	"github.com/vaynedu/hollow/internal/logger"
	"github.com/vaynedu/hollow/internal/middleware"
)

func main() {
	// 初始化配置
	cfg, err := config.NewConfig(".", "conf")
	if err != nil {
		panic(err)
	}

	// 初始化日志
	log, err := logger.InitLogger(cfg)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	// 用户自定义中间件
	userMiddlewares := []gin.HandlerFunc{
		middleware.RequestIDMiddleware(), // 示例 request_id 中间件
	}

	// 创建App选项
	opts := hollow.AppOption{
		Middlewares: userMiddlewares,
		Config:      cfg,
		Logger:      log,
	}

	// 创建 App 实例
	app, err := hollow.NewApp(opts)
	if err != nil {
		panic(err)
	}

	// 注册路由
	app.Engine.GET("/hello", handler.HelloHandler)

	// 启动服务
	go func() {
		if err := app.Start(); err != nil {
			app.Logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// 关闭服务
	app.End()
}
