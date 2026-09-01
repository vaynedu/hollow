package router

import (
	"github.com/vaynedu/hollow"

	"github.com/vaynedu/hollow/example/proto"
	"github.com/vaynedu/hollow/example/service"
)

// Register 注册示例项目的 HTTP 路由。
func Register(app *hollow.App) {
	proto.RegisterUserServiceGinRouter(app.Engine, service.Get())
}
