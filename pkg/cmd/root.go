package cmd

import (
	"fmt"
	"os"

	"sl-cli/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd 代表基础命令，没有子命令时调用它
var rootCmd = &cobra.Command{
	Use:   "sl-cli",
	Short: "sl-cli 是一个极具扩展性的命令行工具",
	Long: `sl-cli 是一个用 Go 编写的命令行工具，
支持通过 YAML 配置动态扩展 RESTful API、Shell 脚本和系统命令。`,
	// 如果有需要，可以在这里运行默认逻辑
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute 将所有子命令添加到 root 命令并设置标志。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	fmt.Println("init()")
	cobra.OnInitialize(initConfig)

	// 定义全局标志，例如配置文件路径
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件 (默认为 $HOME/.sl-cli.yaml)")

	// 这里可以定义本地标志
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig 读取配置文件和环境变量
func initConfig() {
	fmt.Println("initConfig()")
	if cfgFile != "" {
		// 使用标志传入的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找 home 目录
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 在 home 目录下搜索名为 ".sl-cli" 的配置 (无扩展名)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".") // 也搜索当前目录
		viper.SetConfigType("yaml")
		viper.SetConfigName(".sl-cli")
	}

	viper.AutomaticEnv() // 读取匹配的环境变量

	// 如果找到配置文件，则读取它
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// 加载动态命令
	loadDynamicCommands()
}

// loadDynamicCommands 读取配置并构建命令树
func loadDynamicCommands() {
	fmt.Println("loadDynamicCommands")
	var cfg config.Config
	// 将 viper 解析到的数据反序列化到结构体中
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Printf("Error parsing config: %s\n", err)
		return
	}

	// 遍历配置中的根命令，添加到 rootCmd
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
		// 这里的 Run 是核心，目前仅打印调试信息，Step 4 会替换为真实执行器
		Run: func(c *cobra.Command, args []string) {
			fmt.Printf("[Mock Execute] Type: %s\n", cfg.Type)
			fmt.Printf("Command: %s, Args: %v\n", cfg.Name, args)

			if cfg.Type == "http" {
				fmt.Printf("Target URL: %s, Headers: %v\n", cfg.API.URL, cfg.API.Headers)
			} else if cfg.Type == "shell" {
				fmt.Println("Script Content:\n", cfg.Script)
			}
		},
	}

	// 递归处理子命令 (SubCommands)
	for _, subCfg := range cfg.SubCommands {
		subCmd := buildCommand(subCfg)
		cmd.AddCommand(subCmd)
	}

	return cmd
}
