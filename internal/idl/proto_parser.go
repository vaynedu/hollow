package idl

import (
	"errors"
	"io/ioutil"
	"strings"
)

// Service 表示从Protobuf解析出的服务定义
type Service struct {
	Name    string
	Methods []Method
}

// Method 表示服务中的方法定义
type Method struct {
	Name         string
	RequestType  string
	ResponseType string
	HTTPMethod   string
	Path         string
	BindingType  string
}

// ParseProto 解析Protobuf文件并提取服务和方法信息
func ParseProto(protoPath string) (*Service, error) {
	data, err := ioutil.ReadFile(protoPath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	service := &Service{}

	// 简单解析示例，实际实现需要更完善的解析逻辑
	serviceStart := strings.Index(content, "service")
	if serviceStart == -1 {
		return nil, errors.New("no service definition found in proto file")
	}

	// 提取服务名
	serviceEnd := strings.Index(content[serviceStart:], "{")
	if serviceEnd == -1 {
		return nil, errors.New("invalid service definition")
	}

	serviceLine := content[serviceStart : serviceStart+serviceEnd]
	service.Name = strings.TrimSpace(strings.TrimPrefix(serviceLine, "service"))

	// 提取方法
	methodStart := serviceStart + serviceEnd + 1
	methodEnd := strings.Index(content[methodStart:], "}")
	if methodEnd == -1 {
		return nil, errors.New("invalid method definitions")
	}

	methodsContent := content[methodStart : methodStart+methodEnd]
	methodLines := strings.Split(methodsContent, ";")

	for _, line := range methodLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "rpc") {
			method := Method{}
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}

			method.Name = parts[1]
			method.RequestType = strings.Trim(parts[2], "()")
			method.ResponseType = strings.Trim(parts[4], "()")
			method.HTTPMethod = "POST" // 默认POST方法
			method.Path = "/" + strings.ToLower(method.Name)
			method.BindingType = "JSON" // 默认JSON绑定

			service.Methods = append(service.Methods, method)
		}
	}

	if len(service.Methods) == 0 {
		return nil, errors.New("no valid rpc methods found")
	}

	return service, nil
}