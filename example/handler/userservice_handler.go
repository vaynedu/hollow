package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/example/proto"
	"github.com/vaynedu/hollow/example/service"
)


// CreateUserHandler CreateUser 接口处理器
// TODO: 实现业务逻辑
func CreateUserHandler(c *gin.Context) {
	var req proto.CreateUserRequest
	
	// POST/PUT/DELETE 请求：使用 ShouldBindJSON 绑定请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	

	// 调用 service 层
	resp, err := service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	// 设置数据，由responseMiddleware统一处理响应
	c.Set("data", resp)
}

// GetUserHandler GetUser 接口处理器
// TODO: 实现业务逻辑
func GetUserHandler(c *gin.Context) {
	var req proto.GetUserRequest
	
	// GET 请求：使用 ShouldBindQuery 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}
	

	// 调用 service 层
	resp, err := service.GetUser(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	// 设置数据，由responseMiddleware统一处理响应
	c.Set("data", resp)
}

// QueryUsersHandler QueryUsers 接口处理器
// TODO: 实现业务逻辑
func QueryUsersHandler(c *gin.Context) {
	var req proto.QueryUsersRequest
	
	// GET 请求：使用 ShouldBindQuery 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}
	

	// 调用 service 层
	resp, err := service.QueryUsers(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	// 设置数据，由responseMiddleware统一处理响应
	c.Set("data", resp)
}

// UpdateUserHandler UpdateUser 接口处理器
// TODO: 实现业务逻辑
func UpdateUserHandler(c *gin.Context) {
	var req proto.UpdateUserRequest
	
	// POST/PUT/DELETE 请求：使用 ShouldBindJSON 绑定请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	

	// 调用 service 层
	resp, err := service.UpdateUser(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	// 设置数据，由responseMiddleware统一处理响应
	c.Set("data", resp)
}

// DeleteUserHandler DeleteUser 接口处理器
// TODO: 实现业务逻辑
func DeleteUserHandler(c *gin.Context) {
	var req proto.DeleteUserRequest
	
	// POST/PUT/DELETE 请求：使用 ShouldBindJSON 绑定请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	

	// 调用 service 层
	resp, err := service.DeleteUser(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	// 设置数据，由responseMiddleware统一处理响应
	c.Set("data", resp)
}

