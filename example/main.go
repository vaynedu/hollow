package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow"
	"github.com/vaynedu/hollow/example/handler"
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
	// app.Engine.GET("/hello", handler.HelloHandler)
	app.AddRoute(http.MethodGet, "/hello", handler.HelloHandler)

	// 启动服务
	app.Start()

	// 关闭服务
	app.End()
}
