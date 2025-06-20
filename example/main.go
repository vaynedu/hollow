package main

import (
	"net/http"

	"github.com/vaynedu/hollow"
	"github.com/vaynedu/hollow/example/handler"
	"github.com/vaynedu/hollow/internal/middleware"
)

func main() {

	// 创建App选项
	opts := hollow.AppOption{
		AddMiddlewares: []middleware.Middleware{
			middleware.NewRequestIDMiddleware(), // 示例 request_id 中间件
		},
		RemoveMiddlewares: []middleware.Middleware{
			//middleware.NewResponseMiddleware(), // 示例 统一reponse打印
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
