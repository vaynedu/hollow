package generator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultHollowVersion = "v1.0.0"

var projectTemplateWriter = writeTemplate

type ProjectOptions struct {
	Module        string
	Service       string
	HollowVersion string
	HollowPath    string
}

type ProjectConfig struct {
	ProjectName   string
	ModuleName    string
	ServiceName   string
	ProtoName     string
	ProtoPackage  string
	GoVersion     string
	HollowVersion string
	HollowPath    string
}

func newProjectConfig(projectPath string, options ProjectOptions) (ProjectConfig, error) {
	projectName := filepath.Base(filepath.Clean(projectPath))
	moduleName := options.Module
	if moduleName == "" {
		moduleName = projectName
	}

	serviceName := options.Service
	if serviceName == "" {
		serviceName = deriveServiceName(projectName)
	}

	hollowVersion := options.HollowVersion
	if hollowVersion == "" {
		hollowVersion = DefaultHollowVersion
	}

	protoName := deriveProtoName(projectName)
	return ProjectConfig{
		ProjectName:   projectName,
		ModuleName:    moduleName,
		ServiceName:   serviceName,
		ProtoName:     protoName,
		ProtoPackage:  protoName,
		GoVersion:     "1.23.4",
		HollowVersion: hollowVersion,
		HollowPath:    options.HollowPath,
	}, nil
}

func deriveServiceName(projectName string) string {
	parts := projectNameParts(projectName)
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "") + "Service"
}

func deriveProtoName(projectName string) string {
	parts := projectNameParts(projectName)
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
	return strings.Join(parts, "_")
}

func projectNameParts(projectName string) []string {
	projectName = strings.TrimSuffix(projectName, "-server")
	projectName = strings.TrimSuffix(projectName, "_server")
	return strings.FieldsFunc(projectName, func(r rune) bool {
		return r == '-' || r == '_'
	})
}

// InitProject 在一个全新的目标目录中生成固定项目骨架。
func InitProject(projectPath string, options ProjectOptions) error {
	return initProject(projectPath, options, projectTemplateWriter)
}

func initProject(
	projectPath string,
	options ProjectOptions,
	write func(string, string, ProjectConfig) error,
) error {
	config, err := newProjectConfig(projectPath, options)
	if err != nil {
		return err
	}

	if _, err := os.Lstat(projectPath); err == nil {
		return fmt.Errorf("target %q already exists", projectPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check target %q: %w", projectPath, err)
	}

	temporaryPath, err := os.MkdirTemp(
		filepath.Dir(projectPath),
		"."+filepath.Base(projectPath)+"-",
	)
	if err != nil {
		return fmt.Errorf("create temporary project for %q: %w", projectPath, err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(temporaryPath)
		}
	}()

	for _, file := range projectFiles(config) {
		target := filepath.Join(temporaryPath, file.path)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("create directory for %q: %w", target, err)
		}
		if err := write(target, file.template, config); err != nil {
			return err
		}
	}

	if err := os.Rename(temporaryPath, projectPath); err != nil {
		return fmt.Errorf("publish project %q: %w", projectPath, err)
	}
	committed = true
	return nil
}
