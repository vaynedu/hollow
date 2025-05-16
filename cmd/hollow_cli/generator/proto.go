// Package generator hollow/cmd/hollow_cli/generator/proto.go
package generator

import (
	"fmt"
	"hollow/internal/idl" // 自定义IDL解析器
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// GenerateProto 生成HTTP处理代码
func GenerateProto(protoPath string) error {
	// 解析Protobuf文件（使用idl包解析服务和方法）
	service, err := idl.ParseProto(protoPath)
	if err != nil {
		return err
	}

	// 定义模板（生成路由和Handler接口）
	tmpl, err := template.New("handler").Parse(`
package {{.Package}}

import (
	"{{.FrameworkImport}}/hollow"
	"github.com/gin-gonic/gin"
)

// 自动生成的路由注册函数
func Register{{.ServiceName}}Routes(router *gin.RouterGroup, handler {{.ServiceName}}Handler) {
	{{range .Methods}}
	router.{{.HTTPMethod}}("{{.Path}}", func(c *gin.Context) {
		var req {{.RequestType}}
		if err := c.ShouldBind{{.BindingType}}(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		resp, err := handler.{{.MethodName}}(c, &req)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Set("data", resp) // 传递给统一响应中间件
	})
	{{end}}
}

// 业务需实现的接口
type {{.ServiceName}}Handler interface {
	{{range .Methods}}
	{{.MethodName}}(ctx *gin.Context, req *{{.RequestType}}) (*{{.ResponseType}}, error)
	{{end}}
}
`)

	// 生成文件路径
	outputDir := filepath.Join(filepath.Dir(protoPath), "handler")
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, service.Name+"_handler.go")

	// 执行模板渲染
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := map[string]interface{}{
		"Package":         filepath.Base(outputDir),
		"FrameworkImport": "github.com/vaynedu/hollow", // 替换为实际导入路径
		"ServiceName":     service.Name,
		"Methods":         service.Methods,
	}
	return tmpl.Execute(file, data)
}

// GenerateGoFromProto 调用 protoc 工具将 proto 文件生成对应的 go 文件
func GenerateGoFromProto(protoPath string) error {
	// 获取 proto 文件所在目录
	protoDir := filepath.Dir(protoPath)

	// 构建 protoc 命令
	cmd := exec.Command("protoc",
		"--go_out=.",
		"--proto_path="+protoDir,
		filepath.Base(protoPath),
	)

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
