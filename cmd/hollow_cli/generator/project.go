package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type ProjectConfig struct {
	ProjectName      string
	ModuleName       string
	GoVersion        string
	FrameworkVersion string
	HollowPath       string
}

func InitProject(projectName, moduleName string) error {
	if projectName == "" {
		projectName = "myapp"
	}
	if moduleName == "" {
		moduleName = "github.com/example/" + projectName
	}

	// 计算 hollow 框架的相对路径
	// 生成的项目在 hollow/cmd/hollow_cli/<projectName>
	// hollow 框架在 hollow 目录
	// 所以相对路径是 ../../..
	hollowPath := "../../.."

	config := ProjectConfig{
		ProjectName:      projectName,
		ModuleName:       moduleName,
		GoVersion:        "1.23.4",
		FrameworkVersion: "v0.1.0",
		HollowPath:       hollowPath,
	}

	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("创建项目目录失败: %w", err)
	}

	generators := []struct {
		path string
		tmpl string
		data interface{}
	}{
		{filepath.Join(projectName, "go.mod"), goModTemplate, config},
		{filepath.Join(projectName, "main.go"), mainTemplate, config},
		{filepath.Join(projectName, "conf.yaml"), configTemplate, config},
		{filepath.Join(projectName, ".gitignore"), gitignoreTemplate, config},
		{filepath.Join(projectName, "Makefile"), makefileTemplate, config},
		{filepath.Join(projectName, "README.md"), readmeTemplate, config},
		{filepath.Join(projectName, "router", "router.go"), baseRouterTemplate, config},
		{filepath.Join(projectName, "handler", "example.go"), exampleHandlerTemplate, config},
		{filepath.Join(projectName, "service", "example.go"), exampleServiceTemplate, config},
		{filepath.Join(projectName, "proto", "example.proto"), exampleProtoTemplate, config},
	}

	for _, g := range generators {
		dir := filepath.Dir(g.path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
		if err := generateFile(g.path, g.tmpl, g.data); err != nil {
			return fmt.Errorf("生成文件 %s 失败: %w", g.path, err)
		}
	}

	fmt.Printf("✅ Project %s created successfully!\n", projectName)
	fmt.Printf("📁 Project path: ./%s\n", projectName)
	fmt.Println("\n🚀 Quick start:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  make init    # 生成代码 + 安装依赖")
	fmt.Println("  make run     # 启动服务")
	fmt.Println("\n📖 Available commands:")
	fmt.Println("  make proto   # 生成 protobuf 和框架代码")
	fmt.Println("  make deps    # 安装依赖")
	fmt.Println("  make build   # 构建项目")
	fmt.Println("  make test    # 运行测试")

	return nil
}

func generateFile(path, tmpl string, data interface{}) error {
	template, err := template.New("file").Parse(tmpl)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return template.Execute(file, data)
}

var goModTemplate = `module {{.ModuleName}}

go {{.GoVersion}}

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/vaynedu/hollow {{.FrameworkVersion}}
	go.uber.org/zap v1.27.0
)

replace github.com/vaynedu/hollow => {{.HollowPath}}
`

var mainTemplate = `package main

import (
	"{{.ModuleName}}/router"
	"github.com/vaynedu/hollow"
)

func main() {
	opts := hollow.AppOption{
		ConfigPath: ".",
		ConfigName: "conf",
	}

	app, err := hollow.NewApp(opts)
	if err != nil {
		panic(err)
	}

	// 注册所有路由
	router.RegisterRoutes(app)

	app.Start()
	app.End()
}
`

var configTemplate = `host: 127.0.0.1:8080

log:
  level: debug
  output_mode: console
  file: app.log
  max_size: 100
  max_age: 30

db:
  dsn: root:123456@tcp(127.0.0.1:3306)/{{.ProjectName}}?charset=utf8mb4&parseTime=True&loc=Local
  dialect: mysql

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
`

var gitignoreTemplate = `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test
*.test
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Application
*.log
/tmp/
/dist/
/bin/
`

var makefileTemplate = `.PHONY: build run test clean proto deps fmt init

# 默认构建目标
build:
	go build -o bin/{{.ProjectName}} main.go

# 运行服务
run:
	go run main.go

# 运行测试
test:
	go test -v ./...

# 清理构建产物
clean:
	rm -rf bin/

# 获取 proto 目录下的所有 .proto 文件
PROTO_FILES := $(wildcard proto/*.proto)

# 检测 hollow-cli 路径
# 优先使用 PATH 中的 hollow-cli，如果没有则使用相对路径
HOLLOW_CLI := $(shell which hollow-cli 2>/dev/null || echo "../hollow-cli")

# 生成 protobuf 代码和 Hollow 框架代码
proto:
	@echo "🚀 生成 protobuf 代码..."
	@mkdir -p docs
	@if [ -z "$(PROTO_FILES)" ]; then \
		echo "⚠️  未找到 proto 文件，跳过生成"; \
	else \
		protoc $(PROTO_FILES) \
			-I . \
			--proto_path=/usr/local/include/ \
			--proto_path=$$(go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/ \
			--go_out=. --go_opt=paths=source_relative \
			--myhttp_out=. --myhttp_opt=paths=source_relative \
			--openapiv2_out=./docs/; \
		echo "✅ protobuf 代码生成完成"; \
		echo "🚀 生成 Hollow 框架代码..."; \
		echo "   使用 hollow-cli: $(HOLLOW_CLI)"; \
		for proto_file in $(PROTO_FILES); do \
			echo "   处理: $$proto_file"; \
			$(HOLLOW_CLI) proto $$proto_file; \
		done; \
		echo "✅ Hollow 框架代码生成完成"; \
	fi

# 安装依赖
deps:
	go mod tidy

# 格式化代码
fmt:
	go fmt ./...

# 一键初始化项目（生成代码 + 安装依赖）
init: proto deps
	@echo "✅ 项目初始化完成，运行 'make run' 启动服务"`

var readmeTemplate = `# {{.ProjectName}}

A Go project based on [Hollow](https://github.com/vaynedu/hollow).
`

var baseRouterTemplate = `package router

import (
	"github.com/vaynedu/hollow"
)

// RegisterRoutes 注册所有路由
// 在 main.go 中调用此函数注册路由
// 每个服务的路由在对应的 *_router.go 文件中定义
func RegisterRoutes(app *hollow.App) {
	// TODO: 在这里注册所有服务的路由
	// 例如: RegisterExampleServiceRoutes(app.Engine)
}
`

var exampleHandlerTemplate = `package handler

import (
	"net/http"

	"{{.ModuleName}}/service"
	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hecode"
)

type HelloRequest struct {
	Name string ` + "`" + `form:"name" json:"name"` + "`" + `
}

type HelloResponse struct {
	Message string ` + "`" + `json:"message"` + "`" + `
}

func HelloHandler(c *gin.Context) {
	var req HelloRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, hecode.Error(400, err.Error()))
		return
	}

	resp, err := service.Hello(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, hecode.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, hecode.Success(resp))
}
`

var exampleServiceTemplate = `package service

import (
	"context"
	"fmt"
)

func Hello(ctx context.Context, name string) (map[string]interface{}, error) {
	if name == "" {
		name = "World"
	}
	return map[string]interface{}{
		"message": fmt.Sprintf("Hello, %s!", name),
	}, nil
}
`

var exampleProtoTemplate = `syntax = "proto3";

package {{.ProjectName}}.example;

option go_package = "proto/";

import "google/api/annotations.proto";

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
}

service ExampleService {
  rpc SayHello (HelloRequest) returns (HelloResponse) {
    option (google.api.http) = {
      get: "/v1/hello"
    };
  }
}
`
