// Package generator hollow/cmd/hollow_cli/generator/proto.go
package generator

import (
	"fmt"
	"github.com/vaynedu/hollow/internal/idl" // 自定义IDL解析器
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
)

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
		resp, err := handler.{{.Name}}(c, &req)
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
	{{.Name}}(ctx *gin.Context, req *{{.RequestType}}) (*{{.ResponseType}}, error)
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
		"FrameworkImport": "github.com/vaynedu/hollow",
		"ServiceName":     service.Name,
		"Methods":         service.Methods,
	}

	// 为每个方法设置HTTP方法和路径
	for i := range data["Methods"].([]idl.Method) {
		method := data["Methods"].([]idl.Method)[i]
		method.HTTPMethod = inferHTTPMethod(method.Name)
		method.Path = "/" + strings.ToLower(method.Name)
		data["Methods"].([]idl.Method)[i] = method
	}
	return tmpl.Execute(file, data)
}

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
