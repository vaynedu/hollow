package main

import (
	"github.com/spf13/cobra"
	"github.com/vaynedu/hollow/cmd/hollow_cli/generator"
)

var (
	cliVersion  = "dev"
	initProject = generator.InitProject
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:     "hollow-cli",
		Short:   "Hollow 框架项目生成工具",
		Version: cliVersion,
	}
	root.AddCommand(newInitCommand())
	return root
}

func newInitCommand() *cobra.Command {
	var options generator.ProjectOptions
	cmd := &cobra.Command{
		Use:   "init <project>",
		Short: "初始化一个新的 Hollow 项目",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return initProject(args[0], options)
		},
	}
	cmd.Flags().StringVar(&options.Module, "module", "", "覆盖 Go module")
	cmd.Flags().StringVar(&options.Service, "service", "", "覆盖 Proto Service 名")
	cmd.Flags().StringVar(&options.HollowVersion, "hollow-version", generator.DefaultHollowVersion, "覆盖 Hollow 版本")
	cmd.Flags().StringVar(&options.HollowPath, "hollow-path", "", "使用本地 Hollow 路径")
	return cmd
}
