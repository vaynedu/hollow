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

	// 匹配服务定义 - 使用更灵活的方式
	serviceRegex := regexp.MustCompile(`service\s+(\w+)\s*\{`)
	serviceMatches := serviceRegex.FindStringSubmatchIndex(protoContent)
	if len(serviceMatches) < 4 {
		return nil, errors.New("no service definition found in proto file")
	}

	// 提取服务名
	serviceNameMatch := serviceRegex.FindStringSubmatch(protoContent)
	service := &Service{
		Name: serviceNameMatch[1],
	}

	// 找到服务定义的结束位置（匹配大括号）
	matchStart := serviceMatches[0]

	// 从匹配字符串的开始位置向后查找第一个 {
	startIdx := matchStart
	for i := matchStart; i < len(protoContent); i++ {
		if protoContent[i] == '{' {
			startIdx = i + 1
			break
		}
	}

	braceCount := 1
	endIdx := startIdx

	for i := startIdx; i < len(protoContent); i++ {
		if protoContent[i] == '{' {
			braceCount++
		} else if protoContent[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i
				break
			}
		}
	}

	// 提取服务体内容
	serviceBody := protoContent[startIdx:endIdx]

	// 解析 RPC 方法
	methods := parseRPCMethods(serviceBody)
	service.Methods = methods

	if len(service.Methods) == 0 {
		return nil, errors.New("no valid rpc methods found")
	}

	return service, nil
}

// parseRPCMethods 解析服务体中的 RPC 方法
func parseRPCMethods(serviceBody string) []Method {
	var methods []Method
	lines := strings.Split(serviceBody, "\n")

	var currentMethod *Method
	var inOptionBlock bool
	var optionBraceCount int
	var optionBuffer strings.Builder

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// 检测 RPC 方法定义开始
		if strings.HasPrefix(line, "rpc ") {
			// 如果之前有方法，保存它
			if currentMethod != nil {
				methods = append(methods, *currentMethod)
			}

			// 解析 RPC 定义
			// 格式: rpc MethodName (Request) returns (Response) {
			rpcRegex := regexp.MustCompile(`rpc\s+(\w+)\s*\(([\w.]+)\)\s*returns\s*\(([\w.]+)\)\s*\{?`)
			matches := rpcRegex.FindStringSubmatch(line)
			if len(matches) >= 4 {
				methodName := matches[1]
				currentMethod = &Method{
					Name:         methodName,
					RequestType:  matches[2],
					ResponseType: matches[3],
					HTTPMethod:   inferHTTPMethod(methodName),
					Path:         "/" + strings.ToLower(methodName),
					BindingType:  "JSON",
				}
			}

			// 检查是否在同一行开始了 option 块
			if strings.Contains(line, "option") || strings.Contains(line, "{") {
				// 查找本行或后续行的 option 块
				if strings.Contains(line, "option") {
					inOptionBlock = true
					optionBraceCount = 0
					optionBuffer.Reset()
				}
			}
			continue
		}

		// 如果在 option 块中
		if inOptionBlock {
			optionBuffer.WriteString(line + "\n")

			// 计算大括号
			for _, ch := range line {
				if ch == '{' {
					optionBraceCount++
				} else if ch == '}' {
					optionBraceCount--
					if optionBraceCount <= 0 {
						// option 块结束，解析 HTTP 注解
						if currentMethod != nil {
							httpMethod, httpPath := parseHTTPAnnotation(optionBuffer.String())
							if httpMethod != "" {
								currentMethod.HTTPMethod = httpMethod
							}
							if httpPath != "" {
								currentMethod.Path = httpPath
							}
						}
						inOptionBlock = false
						optionBuffer.Reset()
						break
					}
				}
			}
			continue
		}

		// 检测 option 块开始
		if strings.HasPrefix(line, "option") && currentMethod != nil {
			inOptionBlock = true
			optionBraceCount = 0
			optionBuffer.Reset()
			optionBuffer.WriteString(line + "\n")

			// 计算本行的大括号
			for _, ch := range line {
				if ch == '{' {
					optionBraceCount++
				} else if ch == '}' {
					optionBraceCount--
					if optionBraceCount <= 0 {
						// 单行 option 块
						httpMethod, httpPath := parseHTTPAnnotation(optionBuffer.String())
						if httpMethod != "" {
							currentMethod.HTTPMethod = httpMethod
						}
						if httpPath != "" {
							currentMethod.Path = httpPath
						}
						inOptionBlock = false
						optionBuffer.Reset()
					}
				}
			}
			continue
		}

		// 检测 RPC 方法结束（闭合大括号）
		if line == "}" && currentMethod != nil && !inOptionBlock {
			methods = append(methods, *currentMethod)
			currentMethod = nil
		}
	}

	// 处理最后一个方法
	if currentMethod != nil {
		methods = append(methods, *currentMethod)
	}

	return methods
}

func inferHTTPMethod(methodName string) string {
	switch {
	case strings.HasPrefix(methodName, "Get"), strings.HasPrefix(methodName, "Query"), strings.HasPrefix(methodName, "Lookup"):
		return "GET"
	case strings.HasPrefix(methodName, "Create"), strings.HasPrefix(methodName, "Add"), strings.HasPrefix(methodName, "Post"):
		return "POST"
	case strings.HasPrefix(methodName, "Update"), strings.HasPrefix(methodName, "Put"):
		return "PUT"
	case strings.HasPrefix(methodName, "Delete"), strings.HasPrefix(methodName, "Remove"):
		return "DELETE"
	default:
		return "POST"
	}
}

// parseHTTPAnnotation 解析 HTTP 注解，返回 HTTP 方法和路径
func parseHTTPAnnotation(optionBody string) (string, string) {
	// 匹配 google.api.http option 中的各种 HTTP 方法
	// 支持 get, post, put, delete, patch 等
	patterns := []struct {
		method string
		regex  *regexp.Regexp
	}{
		{"GET", regexp.MustCompile(`get:\s*"([^"]+)"`)},
		{"POST", regexp.MustCompile(`post:\s*"([^"]+)"`)},
		{"PUT", regexp.MustCompile(`put:\s*"([^"]+)"`)},
		{"DELETE", regexp.MustCompile(`delete:\s*"([^"]+)"`)},
		{"PATCH", regexp.MustCompile(`patch:\s*"([^"]+)"`)},
	}

	for _, p := range patterns {
		matches := p.regex.FindStringSubmatch(optionBody)
		if len(matches) > 1 {
			return p.method, matches[1]
		}
	}

	return "", ""
}
