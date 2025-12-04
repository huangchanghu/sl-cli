package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd 定义了 version 命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本信息",
	Long:  `显示当前 sl-cli 的版本号、构建时间和 Go 运行时版本`,
	Run: func(cmd *cobra.Command, args []string) {
		// 这里可以写任意复杂的 Go 逻辑
		fmt.Printf("sl-cli version: v0.1.0\n")
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	// 【关键】将 versionCmd 添加到 rootCmd
	// 由于它们都在 cmd 包下，version.go 可以直接访问 root.go 中的 rootCmd 变量
	rootCmd.AddCommand(versionCmd)
}
