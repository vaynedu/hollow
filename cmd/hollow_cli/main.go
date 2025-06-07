package main

import (
	"github.com/spf13/cobra"
	"github.com/vaynedu/hollow/cmd/hollow_cli/generator"
	"log"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "hollow-cli",
		Short: "Hollow框架代码生成工具",
	}

	var protoImportPaths []string
	var protoCmd = &cobra.Command{
		Use:   "proto [proto文件路径]",
		Short: "从Protobuf文件生成代码",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			protoPath := args[0]

			// 1. 生成Go代码
			if err := generator.GenerateGoFromProto(protoPath, protoImportPaths); err != nil {
				log.Fatalf("生成Go代码失败: %v", err)
			}

			// 2. 生成HTTP处理代码
			if err := generator.GenerateProto(protoPath); err != nil {
				log.Fatalf("生成HTTP处理代码失败: %v", err)
			}

			log.Println("代码生成成功!")
		},
	}

	protoCmd.Flags().StringSliceVarP(&protoImportPaths, "proto_path", "I", []string{}, "Protobuf文件引用路径")

	rootCmd.AddCommand(protoCmd)
	rootCmd.Execute()
}
