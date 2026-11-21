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
	"github.com/subosito/gotenv"
)

var cfgFile string

// builtinCommands 记录加载动态命令前已注册的原生命令名，供 config check 识别同名合并
var builtinCommands = map[string]bool{}

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

// stripConfigFlag 去掉参数中第一个 --config <path> / --config=<path>（与 preParseConfigFlag 的识别规则一致）
func stripConfigFlag(args []string) []string {
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			return append(append([]string{}, args[:i]...), args[i+2:]...)
		}
		if strings.HasPrefix(arg, "--config=") {
			return append(append([]string{}, args[:i]...), args[i+1:]...)
		}
	}
	return args
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

		// Only canonical path: ~/.config/sl-cli/sl-cli.yaml
		defaultConfigPath := filepath.Join(home, ".config", "sl-cli", "sl-cli.yaml")
		if _, err := os.Stat(defaultConfigPath); err == nil {
			viper.SetConfigFile(defaultConfigPath)
		}

		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	_ = viper.ReadInConfig()

	loadDotEnv()
}

// loadDotEnv 从配置文件所在目录加载 .env 到进程环境变量。
// 用于把敏感值（token、idcode 等）与 YAML 分离，避免明文写在配置里。
// 使用 gotenv.Load（非覆盖式）：shell 中已导出的同名环境变量优先，.env 仅补充缺失项。
func loadDotEnv() {
	var dir string
	if used := viper.ConfigFileUsed(); used != "" {
		dir = filepath.Dir(used)
	} else if home, err := os.UserHomeDir(); err == nil {
		dir = filepath.Join(home, ".config", "sl-cli")
	}
	if dir == "" {
		return
	}

	envPath := filepath.Join(dir, ".env")
	if _, err := os.Stat(envPath); err != nil {
		return // .env 不存在则跳过
	}
	_ = gotenv.Load(envPath)
}

// loadDynamicCommands 读取配置并构建命令树
func loadDynamicCommands() {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		// 写到 stderr：__complete 会把 stdout 每行当作补全候选
		fmt.Fprintf(os.Stderr, "Error loading config: %s\n", err)
		return
	}

	for _, c := range rootCmd.Commands() {
		builtinCommands[c.Name()] = true
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

	// 配置了补全脚本时，按 TAB 由 cobra 的 __complete 回调执行脚本生成候选
	if comp := cfg.Completion; comp != nil && comp.Script != "" {
		cmd.ValidArgsFunction = func(c *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			directive := cobra.ShellCompDirectiveNoFileComp
			if comp.FileFallback {
				directive = cobra.ShellCompDirectiveDefault
			}
			if comp.NoSpace {
				directive |= cobra.ShellCompDirectiveNoSpace
			}

			// 禁用 flag 解析时 cobra 不会剔除 sl-cli 自身的 --config，需手动去掉
			if c.DisableFlagParsing {
				args = stripConfigFlag(args)
			}

			candidates, err := executor.Complete(comp, c.CommandPath(), args, toComplete, vars)
			if err != nil {
				// 补全失败不能打断用户输入，只写调试日志 (BASH_COMP_DEBUG_FILE)
				cobra.CompDebugln(err.Error(), true)
				return nil, directive
			}
			return candidates, directive
		}
	}

	for _, subCfg := range cfg.SubCommands {
		subCmd := buildCommand(subCfg, vars)
		cmd.AddCommand(subCmd)
	}

	return cmd
}
