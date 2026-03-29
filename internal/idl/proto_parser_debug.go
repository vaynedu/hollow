package idl

import (
	"fmt"
)

// ParseProtoDebug 解析Protobuf文件并打印调试信息
func ParseProtoDebug(protoPath string) (*Service, error) {
	service, err := ParseProto(protoPath)
	if err != nil {
		return nil, err
	}
	
	fmt.Printf("Service: %s\n", service.Name)
	fmt.Printf("Methods count: %d\n", len(service.Methods))
	for i, m := range service.Methods {
		fmt.Printf("  %d. %s (%s) -> %s [%s %s]\n", i+1, m.Name, m.RequestType, m.ResponseType, m.HTTPMethod, m.Path)
	}
	
	return service, nil
}
