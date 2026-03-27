package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/vaynedu/hollow/cmd/hollow_cli/generator"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "hollow-cli",
		Short: "Hollow框架代码生成工具",
		Long:  `hollow-cli 是一个用于快速创建 Hollow 框架项目和生成代码的工具。`,
	}

	// init 命令 - 初始化新项目
	var projectName, moduleName string
	var initCmd = &cobra.Command{
		Use:   "init [项目名称]",
		Short: "初始化一个新的 Hollow 项目",
		Long:  `创建一个新的 Hollow 项目，包含完整的目录结构、配置文件和示例代码。`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				projectName = args[0]
			}
			if err := generator.InitProject(projectName, moduleName); err != nil {
				log.Fatalf("初始化项目失败: %v", err)
			}
		},
	}
	initCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Go 模块名称 (例如: github.com/username/project)")

	// proto 命令 - 从 proto 文件生成代码
	var protoImportPaths []string
	var protoCmd = &cobra.Command{
		Use:   "proto [proto文件路径]",
		Short: "从 Protobuf 文件生成代码",
		Long:  `解析 Protobuf 文件并生成对应的 Go 代码、Handler 和 Service 模板。`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			protoPath := args[0]

			// 1. 生成 Go 代码 (protoc)
			if err := generator.GenerateGoFromProto(protoPath, protoImportPaths); err != nil {
				log.Printf("生成 Go 代码失败: %v", err)
				log.Println("跳过 protoc 代码生成，继续生成 Handler 和 Service...")
			}

			// 2. 生成 Handler 和 Service 模板
			if err := generator.GenerateProto(protoPath); err != nil {
				log.Fatalf("生成 Handler 和 Service 失败: %v", err)
			}

			log.Println("✅ 代码生成成功!")
		},
	}
	protoCmd.Flags().StringSliceVarP(&protoImportPaths, "proto_path", "I", []string{}, "Protobuf 文件引用路径")

	// 添加子命令
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(protoCmd)

	// 执行
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
