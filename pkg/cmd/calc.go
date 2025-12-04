package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// calcCmd 演示原生参数处理
var calcCmd = &cobra.Command{
	Use:   "add [n1] [n2]",
	Short: "计算两个数字之和 (原生 Go 实现)",
	Args:  cobra.ExactArgs(2), // 强制要求两个参数
	Run: func(cmd *cobra.Command, args []string) {
		n1, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("参数错误: %s 不是有效数字\n", args[0])
			return
		}
		n2, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("参数错误: %s 不是有效数字\n", args[1])
			return
		}

		sum := n1 + n2
		fmt.Printf("%d + %d = %d\n", n1, n2, sum)
	},
}

func init() {
	rootCmd.AddCommand(calcCmd)
}
