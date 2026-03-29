package service

import (
	"context"

	"github.com/vaynedu/hollow/example/proto"
)

// UserServiceService UserService 服务实现
type UserServiceService struct{}

// NewUserServiceService 创建服务实例
func NewUserServiceService() *UserServiceService {
	return &UserServiceService{}
}

// 全局服务实例
var userServiceService = NewUserServiceService()


// CreateUser CreateUser business logic (全局函数，供Handler调用)
func CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	return userServiceService.CreateUser(ctx, req)
}

// CreateUser CreateUser business logic
// TODO: Implement specific business logic
func (s *UserServiceService) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	// TODO: Implement business logic here
	// 默认返回空响应，业务同学根据实际需求修改
	return &proto.CreateUserResponse{}, nil
}

// GetUser GetUser business logic (全局函数，供Handler调用)
func GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	return userServiceService.GetUser(ctx, req)
}

// GetUser GetUser business logic
// TODO: Implement specific business logic
func (s *UserServiceService) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	// TODO: Implement business logic here
	// 默认返回空响应，业务同学根据实际需求修改
	return &proto.GetUserResponse{}, nil
}

// QueryUsers QueryUsers business logic (全局函数，供Handler调用)
func QueryUsers(ctx context.Context, req *proto.QueryUsersRequest) (*proto.QueryUsersResponse, error) {
	return userServiceService.QueryUsers(ctx, req)
}

// QueryUsers QueryUsers business logic
// TODO: Implement specific business logic
func (s *UserServiceService) QueryUsers(ctx context.Context, req *proto.QueryUsersRequest) (*proto.QueryUsersResponse, error) {
	// TODO: Implement business logic here
	// 默认返回空响应，业务同学根据实际需求修改
	return &proto.QueryUsersResponse{}, nil
}

// UpdateUser UpdateUser business logic (全局函数，供Handler调用)
func UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	return userServiceService.UpdateUser(ctx, req)
}

// UpdateUser UpdateUser business logic
// TODO: Implement specific business logic
func (s *UserServiceService) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	// TODO: Implement business logic here
	// 默认返回空响应，业务同学根据实际需求修改
	return &proto.UpdateUserResponse{}, nil
}

// DeleteUser DeleteUser business logic (全局函数，供Handler调用)
func DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	return userServiceService.DeleteUser(ctx, req)
}

// DeleteUser DeleteUser business logic
// TODO: Implement specific business logic
func (s *UserServiceService) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	// TODO: Implement business logic here
	// 默认返回空响应，业务同学根据实际需求修改
	return &proto.DeleteUserResponse{}, nil
}

