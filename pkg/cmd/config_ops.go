package cmd

import (
	"fmt"
	"os"

	"sl-cli/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd 是配置相关的父命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理工具",
}

// checkCmd 用于检查配置文件的语法和逻辑有效性
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "检查配置文件的语法和逻辑错误",
	Run: func(cmd *cobra.Command, args []string) {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			fmt.Println("❌ Config file not found at $HOME/.config/sl-cli/sl-cli.yaml.")
			fmt.Println("Reinstall via 'make install' to provision the default config.")
			os.Exit(1)
		}

		fmt.Printf("✅ Config file found: %s\n", configFile)

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Printf("❌ Config Loading Error: %s\n", err)
			os.Exit(1)
		}

		errCount := 0
		for i, c := range cfg.Commands {
			cmdName := c.Name
			if cmdName == "" {
				cmdName = fmt.Sprintf("Command#%d", i+1)
			}
			errCount += validateCommand(c, cmdName)
		}

		if errCount > 0 {
			fmt.Printf("\nFound %d errors in configuration.\n", errCount)
			os.Exit(1)
		}

		fmt.Println("✅ Configuration is valid! All systems go.")
	},
}

// validateCommand 递归校验命令配置
// c: 当前命令配置
// path: 命令路径面包屑，例如 "dev -> info"
func validateCommand(c config.CommandConfig, path string) int {
	errs := 0

	// 1. 基础校验：Name 必须存在
	if c.Name == "" {
		fmt.Printf("❌ Error in [%s]: 'name' is required.\n", path)
		errs++
	}

	// 2. 结构校验：必须是 "有效的功能命令" 或者 "包含子命令的组"
	if c.Type == "" && len(c.SubCommands) == 0 {
		fmt.Printf("❌ Error in [%s]: Must specify 'type' (http/shell/system) OR have 'subcommands'.\n", path)
		errs++
	}

	// 3. 类型校验
	if c.Type != "" {
		validTypes := map[string]bool{"http": true, "shell": true, "system": true}
		if !validTypes[c.Type] {
			fmt.Printf("❌ Error in [%s]: Invalid type '%s'. Must be http, shell, or system.\n", path, c.Type)
			errs++
		}

		switch c.Type {
		case "http":
			if c.API.URL == "" {
				fmt.Printf("❌ Error in [%s]: Type is http but 'api.url' is missing.\n", path)
				errs++
			}
			for idx, p := range c.API.Pipes {
				if p.Command == "" {
					fmt.Printf("❌ Error in [%s]: Pipe #%d missing 'command'.\n", path, idx+1)
					errs++
				}
			}
		case "shell":
			if c.Script == "" {
				fmt.Printf("❌ Error in [%s]: Type is shell but 'script' is missing.\n", path)
				errs++
			}
		case "system":
			if c.Command == "" {
				fmt.Printf("❌ Error in [%s]: Type is system but 'command' is missing.\n", path)
				errs++
			}
		}
	}

	// 4. 递归校验子命令
	for _, sub := range c.SubCommands {
		subPath := path + " -> " + sub.Name
		if sub.Name == "" {
			subPath = path + " -> [Unnamed]"
		}
		errs += validateCommand(sub, subPath)
	}

	return errs
}

func init() {
	configCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(configCmd)
}
