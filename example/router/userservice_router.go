package router

import (
	"github.com/vaynedu/hollow/example/handler"
	"github.com/gin-gonic/gin"
)

// RegisterUserServiceRoutes 注册 UserService 服务路由
func RegisterUserServiceRoutes(r *gin.Engine) {

	// CreateUser - POST /v1/users
	r.POST("/v1/users", handler.CreateUserHandler)

	// GetUser - GET /v1/users/{id}
	r.GET("/v1/users/{id}", handler.GetUserHandler)

	// QueryUsers - GET /v1/users
	r.GET("/v1/users", handler.QueryUsersHandler)

	// UpdateUser - PUT /v1/users/{id}
	r.PUT("/v1/users/{id}", handler.UpdateUserHandler)

	// DeleteUser - DELETE /v1/users/{id}
	r.DELETE("/v1/users/{id}", handler.DeleteUserHandler)

}
