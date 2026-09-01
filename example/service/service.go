package service

import "github.com/vaynedu/hollow/example/proto"

// Service 实现示例项目的 Proto Service。
type Service struct {
	proto.UnimplementedUserServiceService
}

var instance = new(Service)

// Get 返回示例项目唯一的 Service 实例。
func Get() *Service {
	return instance
}
