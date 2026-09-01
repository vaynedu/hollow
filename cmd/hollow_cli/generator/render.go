package generator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templateFiles embed.FS

type projectFile struct {
	path     string
	template string
}

func projectFiles(config ProjectConfig) []projectFile {
	return []projectFile{
		{path: ".gitignore", template: "gitignore.tmpl"},
		{path: "README.md", template: "README.md.tmpl"},
		{path: "go.mod", template: "go.mod.tmpl"},
		{path: "main.go", template: "main.go.tmpl"},
		{path: "conf.yaml", template: "conf.yaml.tmpl"},
		{path: "Makefile", template: "Makefile.tmpl"},
		{path: "config/config.go", template: "config.go.tmpl"},
		{path: filepath.Join("proto", config.ProtoName+".proto"), template: "proto.tmpl"},
		{path: "router/router.go", template: "router.go.tmpl"},
		{path: "service/service.go", template: "service.go.tmpl"},
	}
}

func writeTemplate(targetPath, templateName string, config ProjectConfig) error {
	templatePath := filepath.ToSlash(filepath.Join("templates", templateName))
	source, err := templateFiles.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("render %q: read template %q: %w", targetPath, templateName, err)
	}

	tmpl, err := template.New(templateName).Option("missingkey=error").Parse(string(source))
	if err != nil {
		return fmt.Errorf("render %q: parse template %q: %w", targetPath, templateName, err)
	}
	var content bytes.Buffer
	if err := tmpl.Execute(&content, config); err != nil {
		return fmt.Errorf("render %q: execute template %q: %w", targetPath, templateName, err)
	}

	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("write %q: %w", targetPath, err)
	}
	if _, err := file.Write(content.Bytes()); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %q: %w", targetPath, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %q: %w", targetPath, err)
	}
	return nil
}
