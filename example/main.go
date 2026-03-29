package main

import (
	"github.com/vaynedu/hollow"
	"github.com/vaynedu/hollow/example/router"
)

func main() {
	// 创建App选项
	opts := hollow.AppOption{
		ConfigPath: ".",
		ConfigName: "conf",
	}

	// 创建 App 实例
	app, err := hollow.NewApp(opts)
	if err != nil {
		panic(err)
	}

	// 注册路由
	router.RegisterRoutes(app)

	// 启动服务
	app.Start()

	// 关闭服务
	app.End()
}
