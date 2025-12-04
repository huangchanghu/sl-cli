package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// genManCmd 用于生成 Man pages
var genManCmd = &cobra.Command{
	Use:    "gen-man [output-dir]",
	Short:  "生成 Man pages 文档",
	Hidden: true, // 这是一个构建工具命令，不需要对普通用户显示
	Args:   cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 默认输出目录
		outDir := "./man1"
		if len(args) > 0 {
			outDir = args[0]
		}

		// 确保目录存在
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}

		// 设置 Man page 的头部信息
		header := &doc.GenManHeader{
			Title:   "SL-CLI",
			Section: "1",
			Source:  "Sl-Cli Auto Generated",
			Manual:  "Sl-Cli Manual",
		}

		// 生成文档树
		// 注意：这会为 rootCmd 及其所有子命令生成单独的文件
		err := doc.GenManTree(rootCmd, header, outDir)
		if err != nil {
			return err
		}

		fmt.Printf("Man pages generated in: %s\n", outDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genManCmd)
}
