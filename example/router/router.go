package router

import (
	"github.com/vaynedu/hollow"
)

// RegisterRoutes 注册所有路由
// 在 main.go 中调用此函数注册路由
func RegisterRoutes(app *hollow.App) {
	// 注册 UserService 服务路由
	RegisterUserServiceRoutes(app.Engine)
}
