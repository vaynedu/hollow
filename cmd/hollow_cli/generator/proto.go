// Package generator hollow/cmd/hollow_cli/generator/proto.go
package generator

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vaynedu/hollow/internal/idl"
)

// ProtoGenerator 代码生成器配置
type ProtoGenerator struct {
	ProtoPath       string
	OutputDir       string
	ModuleName      string
	FrameworkImport string
}

// GenerateProto 生成HTTP处理代码
func GenerateProto(protoPath string) error {
	// 替换路径中的 ~ 并转换路径
	convertedPath, err := replaceTildeAndConvertPath(protoPath)
	if err != nil {
		return err
	}
	protoPath = convertedPath

	// 解析Protobuf文件（使用idl包解析服务和方法）
	service, err := idl.ParseProto(protoPath)
	if err != nil {
		return err
	}

	// 获取项目模块名
	moduleName := detectModuleName(filepath.Dir(protoPath))

	gen := &ProtoGenerator{
		ProtoPath:       protoPath,
		OutputDir:       filepath.Dir(protoPath),
		ModuleName:      moduleName,
		FrameworkImport: "github.com/vaynedu/hollow",
	}

	// 生成 Handler 文件
	if err := gen.generateHandler(service); err != nil {
		return fmt.Errorf("生成 handler 失败: %w", err)
	}

	// 生成 Service 文件
	if err := gen.generateService(service); err != nil {
		return fmt.Errorf("生成 service 失败: %w", err)
	}

	// 生成 Router 注册文件
	if err := gen.generateRouter(service); err != nil {
		return fmt.Errorf("生成 router 失败: %w", err)
	}

	return nil
}

// 检测项目模块名
func detectModuleName(projectDir string) string {
	// 尝试读取 go.mod 文件（从 proto 目录向上查找）
	for i := 0; i < 3; i++ {
		goModPath := filepath.Join(projectDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			content, err := os.ReadFile(goModPath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "module ") {
						return strings.TrimSpace(strings.TrimPrefix(line, "module"))
					}
				}
			}
			break
		}
		// 向上一级目录查找
		parentDir := filepath.Dir(projectDir)
		if parentDir == projectDir {
			break
		}
		projectDir = parentDir
	}
	return "github.com/example/project"
}

// 生成 Handler 文件
func (g *ProtoGenerator) generateHandler(service *idl.Service) error {
	handlerDir := filepath.Join(g.OutputDir, "..", "handler")
	if err := os.MkdirAll(handlerDir, 0755); err != nil {
		return err
	}

	outputPath := filepath.Join(handlerDir, strings.ToLower(service.Name)+"_handler.go")

	// 检查文件是否已存在，如果存在则不覆盖
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("⚠️  Handler 文件已存在，跳过生成: %s\n", outputPath)
		return nil
	}

	tmpl, err := template.New("handler").Parse(handlerTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := map[string]interface{}{
		"Package":         "handler",
		"ModuleName":      g.ModuleName,
		"FrameworkImport": g.FrameworkImport,
		"ServiceName":     service.Name,
		"Methods":         service.Methods,
	}

	return tmpl.Execute(file, data)
}

// 生成 Service 文件
func (g *ProtoGenerator) generateService(service *idl.Service) error {
	serviceDir := filepath.Join(g.OutputDir, "..", "service")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return err
	}

	outputPath := filepath.Join(serviceDir, strings.ToLower(service.Name)+"_service.go")

	// 检查文件是否已存在，如果存在则不覆盖
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("⚠️  Service 文件已存在，跳过生成: %s\n", outputPath)
		return nil
	}

	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := map[string]interface{}{
		"Package":     "service",
		"ModuleName":  g.ModuleName,
		"ServiceName": service.Name,
		"Methods":     service.Methods,
	}

	return tmpl.Execute(file, data)
}

// 生成 Router 注册文件
func (g *ProtoGenerator) generateRouter(service *idl.Service) error {
	routerDir := filepath.Join(g.OutputDir, "..", "router")
	if err := os.MkdirAll(routerDir, 0755); err != nil {
		return err
	}

	outputPath := filepath.Join(routerDir, strings.ToLower(service.Name)+"_router.go")

	// Router 文件每次重新生成（因为只是注册逻辑）
	tmpl, err := template.New("router").Parse(routerTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 处理方法信息，添加 HTTP 方法和路径
	type MethodInfo struct {
		idl.Method
		HandlerName string
	}

	var methods []MethodInfo
	for _, m := range service.Methods {
		methods = append(methods, MethodInfo{
			Method:      m,
			HandlerName: m.Name + "Handler",
		})
	}

	data := map[string]interface{}{
		"Package":         "router",
		"ModuleName":      g.ModuleName,
		"FrameworkImport": g.FrameworkImport,
		"ServiceName":     service.Name,
		"ServiceVar":      strings.ToLower(service.Name[:1]) + service.Name[1:],
		"Methods":         methods,
	}

	return tmpl.Execute(file, data)
}

// Handler 模板
var handlerTemplate = `package handler

import (
	"net/http"

	"{{.ModuleName}}/service"
	"github.com/gin-gonic/gin"
	"{{.FrameworkImport}}/pkg/hecode"
)

{{range .Methods}}
// {{.Name}}Handler {{.Name}} 接口处理器
// TODO: 实现业务逻辑
func {{.Name}}Handler(c *gin.Context) {
	var req {{.RequestType}}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, hecode.Error(400, err.Error()))
		return
	}

	// 调用 service 层
	resp, err := service.{{.Name}}(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, hecode.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, hecode.Success(resp))
}
{{end}}
`

// Service 模板
var serviceTemplate = `package service

import (
	"context"
	"fmt"

	"{{.ModuleName}}/proto"
)

// {{.ServiceName}}Service {{.ServiceName}} 服务实现
type {{.ServiceName}}Service struct{}

// New{{.ServiceName}}Service 创建服务实例
func New{{.ServiceName}}Service() *{{.ServiceName}}Service {
	return &{{.ServiceName}}Service{}
}

{{range .Methods}}
// {{.Name}} {{.Name}} business logic
// TODO: Implement specific business logic
func (s *{{$.ServiceName}}Service) {{.Name}}(ctx context.Context, req *proto.{{.RequestType}}) (*proto.{{.ResponseType}}, error) {
	// TODO: Implement business logic here
	return &proto.{{.ResponseType}}{
		// Fill response fields
	}, fmt.Errorf("not implemented")
}
{{end}}
`

// Router 模板
var routerTemplate = `package router

import (
	"{{.ModuleName}}/handler"
	"{{.ModuleName}}/service"
	"github.com/gin-gonic/gin"
)

// Register{{.ServiceName}}Routes 注册 {{.ServiceName}} 服务路由
// 在 main.go 中调用此函数注册路由
func Register{{.ServiceName}}Routes(r *gin.RouterGroup) {
	// 创建服务实例
	{{.ServiceVar}}Service := service.New{{.ServiceName}}Service()
	_ = {{.ServiceVar}}Service // 使用服务实例

{{range .Methods}}
	// {{.Name}} - {{.HTTPMethod}} {{.Path}}
	r.{{.HTTPMethod}}("{{.Path}}", handler.{{.HandlerName}})
{{end}}
}
`

// 替换路径中的 ~ 为用户主目录，并转换 Unix 风格路径为 Windows 风格
func replaceTildeAndConvertPath(pathStr string) (string, error) {
	if strings.HasPrefix(pathStr, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		homeDir := usr.HomeDir
		pathStr = strings.Replace(pathStr, "~", homeDir, 1)
	}
	// 转换 Unix 风格路径为 Windows 风格
	if strings.HasPrefix(pathStr, "/c/") {
		pathStr = "c:" + strings.TrimPrefix(pathStr, "/c")
	}
	pathStr = filepath.FromSlash(pathStr)
	return pathStr, nil
}

// GenerateGoFromProto 调用 protoc 工具将 proto 文件生成对应的 go 文件
func GenerateGoFromProto(protoPath string, protoImportPaths []string) error {
	// 替换路径中的 ~ 并转换路径
	protoPath, err := replaceTildeAndConvertPath(protoPath)
	if err != nil {
		return err
	}

	// 获取 proto 文件所在目录
	protoDir := filepath.Dir(protoPath)

	// 构建 protoc 命令参数
	cmdArgs := []string{
		"--go_out=.",
		"--proto_path=" + protoDir,
	}

	// 替换导入路径中的 ~ 并转换路径
	for i, p := range protoImportPaths {
		p, err = replaceTildeAndConvertPath(p)
		if err != nil {
			return err
		}
		protoImportPaths[i] = p
		cmdArgs = append(cmdArgs, "--proto_path="+p)
	}

	cmdArgs = append(cmdArgs, filepath.Base(protoPath))

	cmd := exec.Command("protoc", cmdArgs...)

	// 设置命令执行目录为 proto 文件所在目录
	cmd.Dir = protoDir

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 由于 errors 未定义，使用标准库中的 fmt.Errorf 替代
		return fmt.Errorf("%s: %w", string(output), err)
	}

	return nil
}
