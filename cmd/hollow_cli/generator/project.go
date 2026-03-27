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
}

func InitProject(projectName, moduleName string) error {
	if projectName == "" {
		projectName = "myapp"
	}
	if moduleName == "" {
		moduleName = "github.com/example/" + projectName
	}

	config := ProjectConfig{
		ProjectName:      projectName,
		ModuleName:       moduleName,
		GoVersion:        "1.23.4",
		FrameworkVersion: "v0.1.0",
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

	fmt.Printf("Project %s created successfully!\n", projectName)
	fmt.Printf("Project path: ./%s\n", projectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  go run main.go")
	fmt.Println("\nGenerate proto code:")
	fmt.Println("  hollow-cli proto proto/example.proto")

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
`

var mainTemplate = `package main

import (
	"net/http"

	"{{.ModuleName}}/handler"
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

	app.AddRoute(http.MethodGet, "/hello", handler.HelloHandler)

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

var makefileTemplate = `.PHONY: build run test clean proto

build:
	go build -o bin/{{.ProjectName}} main.go

run:
	go run main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

proto:
	hollow-cli proto proto/example.proto

deps:
	go mod tidy

fmt:
	go fmt ./...
`

var readmeTemplate = `# {{.ProjectName}}

A Go project based on [Hollow](https://github.com/vaynedu/hollow).
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
