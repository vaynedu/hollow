package idl

import (
	"bufio"
	"errors"
	"os"
	"regexp"
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
	file, err := os.Open(protoPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var content strings.Builder
	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	protoContent := content.String()

	// 移除单行注释
	reSingleLineComment := regexp.MustCompile(`//.*`)
	protoContent = reSingleLineComment.ReplaceAllString(protoContent, "")

	// 移除多行注释
	reMultiLineComment := regexp.MustCompile(`/\*.*?\*/`)
	protoContent = reMultiLineComment.ReplaceAllString(protoContent, "")

	// 匹配服务定义
	serviceRegex := regexp.MustCompile(`service\s+(\w+)\s*{([^}]*)}`)
	serviceMatches := serviceRegex.FindStringSubmatch(protoContent)
	if len(serviceMatches) < 3 {
		return nil, errors.New("no service definition found in proto file")
	}

	service := &Service{
		Name: serviceMatches[1],
	}

	// 匹配 RPC 方法，修改正则表达式以支持包含 option 字段的 rpc 定义
	methodRegex := regexp.MustCompile(`rpc\s+(\w+)\s*\(([\w.]+)\)\s*returns\s*\(([\w.]+)\)\s*(?:\{|option)`)
	methodMatches := methodRegex.FindAllStringSubmatch(serviceMatches[2], -1)
	for _, match := range methodMatches {
		if len(match) != 4 {
			continue
		}

		method := Method{
			Name:         match[1],
			RequestType:  match[2],
			ResponseType: match[3],
			HTTPMethod:   inferHTTPMethod(match[1]),
			Path:         "/" + strings.ToLower(match[1]),
			BindingType:  "JSON",
		}
		service.Methods = append(service.Methods, method)
	}

	if len(service.Methods) == 0 {
		return nil, errors.New("no valid rpc methods found")
	}

	return service, nil
}

func inferHTTPMethod(methodName string) string {
	switch {
	case strings.HasPrefix(methodName, "Get"), strings.HasPrefix(methodName, "Query"):
		return "GET"
	case strings.HasPrefix(methodName, "Create"), strings.HasPrefix(methodName, "Add"):
		return "POST"
	case strings.HasPrefix(methodName, "Update"):
		return "PUT"
	case strings.HasPrefix(methodName, "Delete"):
		return "DELETE"
	default:
		return "POST"
	}
}
