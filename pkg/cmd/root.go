package cmd

import (
	"fmt"
	"os"

	"sl-cli/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd 代表基础命令
var rootCmd = &cobra.Command{
	Use:   "sl-cli",
	Short: "sl-cli 是一个极具扩展性的命令行工具",
	Long: `sl-cli 是一个用 Go 编写的命令行工具，
支持通过 YAML 配置动态扩展 RESTful API、Shell 脚本和系统命令。`,
}

// Execute 是主入口
func Execute() {
	// 1. 【核心修正】在执行命令前，主动加载配置和动态命令
	// 注意：此时 Cobra 还没解析命令行参数，所以 --config 标志暂时无法生效
	// 它会优先读取默认路径（当前目录或 Home 目录）下的配置文件
	initConfig()
	loadDynamicCommands()

	// 2. 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 定义全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件 (默认为 $HOME/.sl-cli.yaml)")
}

// initConfig 读取配置文件和环境变量
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".") // 搜索当前目录
		viper.SetConfigType("yaml")
		viper.SetConfigName("sl-cli") // 兼容 sl-cli.yaml 或 .sl-cli.yaml
	}

	viper.AutomaticEnv()

	// 忽略错误，因为如果没配置文件，我们只运行静态命令即可
	if err := viper.ReadInConfig(); err == nil {
		// 仅在调试时打开，避免干扰正常输出
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// loadDynamicCommands 读取配置并构建命令树
func loadDynamicCommands() {
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		// 配置文件格式错误时提示
		fmt.Printf("Error parsing config: %s\n", err)
		return
	}

	for _, cmdCfg := range cfg.Commands {
		cmd := buildCommand(cmdCfg)
		rootCmd.AddCommand(cmd)
	}
}

// buildCommand 递归构建命令
func buildCommand(cfg config.CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfg.Name,
		Short: cfg.Usage,
		Run: func(c *cobra.Command, args []string) {
			// Step 3 Mock Output
			fmt.Printf("[Mock Execute] Type: %s\n", cfg.Type)
			if cfg.Type == "http" {
				fmt.Printf("Target URL: %s\n", cfg.API.URL)
				fmt.Printf("Headers: %v\n", cfg.API.Headers)
			} else if cfg.Type == "shell" {
				fmt.Println("Script Content:\n", cfg.Script)
			}
		},
	}

	for _, subCfg := range cfg.SubCommands {
		subCmd := buildCommand(subCfg)
		cmd.AddCommand(subCmd)
	}

	return cmd
}
