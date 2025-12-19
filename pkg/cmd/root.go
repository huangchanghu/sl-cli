package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sl-cli/internal/config"
	"sl-cli/internal/executor"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd 代表基础命令
var rootCmd = &cobra.Command{
	Use:   "sl-cli",
	Short: "sl-cli 是一个极具扩展性的命令行工具",
	Long: `sl-cli (Super Link CLI) 是一个高度可扩展的现代命令行工具，旨在成为你日常工作流的“超级粘合剂”。
它采用 Go 原生代码 + YAML 动态配置 的混合驱动模式。你既可以通过编写 Go 代码开发高性能的核心命令，也可以通过修改配置文件瞬间挂载 RESTful API、Shell 脚本或系统命令别名，而无需重新编译。`,
}

// Execute 是主入口
func Execute() {
	preParseConfigFlag() // 解析--config配置文件
	initConfig()
	loadDynamicCommands()

	// 2. 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// 简单的 Flag 预解析器
func preParseConfigFlag() {
	if cfgFile != "" {
		return
	}
	args := os.Args
	for i, arg := range args {
		// 检查 --config /path 格式
		if arg == "--config" && i+1 < len(args) {
			cfgFile = args[i+1]
			break
		}
		// 检查 --config=/path 格式
		if strings.HasPrefix(arg, "--config=") {
			cfgFile = strings.TrimPrefix(arg, "--config=")
			break
		}
	}
}

func init() {
	// 定义全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件 (默认为 $HOME/.config/sl-cli/sl-cli.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// New default path: ~/.config/sl-cli/sl-cli.yaml
		configDir := filepath.Join(home, ".config", "sl-cli")
		defaultConfigPath := filepath.Join(configDir, "sl-cli.yaml")

		// Legacy paths (fallback)
		currentDirConfig := "sl-cli.yaml"
		homeDirDotConfig := filepath.Join(home, ".sl-cli.yaml")

		if _, err := os.Stat(defaultConfigPath); err == nil {
			viper.SetConfigFile(defaultConfigPath)
		} else if _, err := os.Stat(currentDirConfig); err == nil {
			viper.SetConfigFile(currentDirConfig)
		} else if _, err := os.Stat(homeDirDotConfig); err == nil {
			viper.SetConfigFile(homeDirDotConfig)
		} else {
			// Set default search path for viper (though we might not strictly need it if we use LoadConfig)
			viper.AddConfigPath(configDir)
			viper.AddConfigPath(".")
			viper.AddConfigPath(home)
			viper.SetConfigName("sl-cli")
		}

		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	// Locate the file using Viper to get the path, but we might load it manually later
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// loadDynamicCommands 读取配置并构建命令树
func loadDynamicCommands() {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		return
	}

	for _, cmdCfg := range cfg.Commands {
		cmd := buildCommand(cmdCfg, cfg.Vars)
		// 重复添加的命令丢弃，如果有子命令则将子命令追加到已存在命令的字命令中
		deplicated := false
		for _, c := range rootCmd.Commands() {
			if c.Name() != cmd.Name() {
				continue
			}
			deplicated = true
			if cmd.HasSubCommands() {
				c.AddCommand(cmd.Commands()...)
			}
		}
		if !deplicated {
			rootCmd.AddCommand(cmd)
		}
	}
}

// buildCommand 递归构建命令
func buildCommand(cfg config.CommandConfig, vars map[string]string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfg.Name,
		Short: cfg.Usage,
		// DisableFlagParsing: true, // 可选：如果希望由 shell/system 接管所有参数解析，可以开启此项
		Run: func(c *cobra.Command, args []string) {
			if err := executor.Run(cfg, args, vars); err != nil {
				fmt.Printf("Execution failed: %s\n", err)
				os.Exit(1)
			}
		},
	}

	// 对于 system 和 shell 类型，禁用 Cobra 的标志解析
	// 这样 -la 这种参数就会被原样放入 args 切片中，而不是被 Cobra 拦截报错
	if cfg.Type == "system" || cfg.Type == "shell" {
		cmd.DisableFlagParsing = true
	}

	for _, subCfg := range cfg.SubCommands {
		subCmd := buildCommand(subCfg, vars)
		cmd.AddCommand(subCmd)
	}

	return cmd
}
