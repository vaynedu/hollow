package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow"
	"github.com/vaynedu/hollow/example/handler"
	"github.com/vaynedu/hollow/internal/middleware"
)

func main() {

	// 创建App选项
	opts := hollow.AppOption{
		AddMiddlewares: []gin.HandlerFunc{
			middleware.RequestIDMiddleware(), // 示例 request_id 中间件
		},
		RemoveMiddlewares: []gin.HandlerFunc{
			middleware.ResponseMiddleware(), // 示例 同意reponse打印
		},
	}

	// 创建 App 实例
	app, err := hollow.NewApp(opts)
	if err != nil {
		panic(err)
	}

	// 注册路由
	app.AddRoute(http.MethodGet, "/hello", handler.HelloHandler)

	// 启动服务
	app.Start()

	// 关闭服务
	app.End()
}
