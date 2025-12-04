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
	Short: "配置管理工具 (检查、初始化)",
}

// checkCmd 用于检查配置文件的语法和逻辑有效性
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "检查配置文件的语法和逻辑错误",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 尝试查找并读取配置
		// 注意：root.go 中的 initConfig 已经运行过，但我们需要显式获取错误信息
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Println("❌ Config file not found.")
				fmt.Println("Run 'sl-cli config init' to generate one.")
			} else {
				fmt.Printf("❌ YAML Syntax Error: %s\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("✅ Config file found: %s\n", viper.ConfigFileUsed())

		// 2. 尝试解析结构体
		var cfg config.Config
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Printf("❌ Config Parsing Error: %s\n", err)
			os.Exit(1)
		}

		// 3. 逻辑校验 (Validate Logic)
		errCount := 0
		for i, c := range cfg.Commands {
			if c.Name == "" {
				fmt.Printf("❌ Error in command #%d: 'name' is required.\n", i+1)
				errCount++
				continue // 名字都没有，没法继续报错
			}

			// 校验类型
			validTypes := map[string]bool{"http": true, "shell": true, "system": true}
			if !validTypes[c.Type] && len(c.SubCommands) == 0 {
				fmt.Printf("❌ Error in command '%s': Invalid type '%s'. Must be http, shell, or system.\n", c.Name, c.Type)
				errCount++
			}

			// 校验具体字段
			switch c.Type {
			case "http":
				if c.API.URL == "" {
					fmt.Printf("❌ Error in command '%s': Type is http but 'api.url' is missing.\n", c.Name)
					errCount++
				}
			case "shell":
				if c.Script == "" {
					fmt.Printf("❌ Error in command '%s': Type is shell but 'script' is missing.\n", c.Name)
					errCount++
				}
			case "system":
				if c.Command == "" {
					fmt.Printf("❌ Error in command '%s': Type is system but 'command' is missing.\n", c.Name)
					errCount++
				}
			}
		}

		if errCount > 0 {
			fmt.Printf("\nFound %d errors in configuration.\n", errCount)
			os.Exit(1)
		}

		fmt.Println("✅ Configuration is valid! All systems go.")
	},
}

// initCmd 用于生成示例配置文件
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "在当前目录生成默认的 sl-cli.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		filename := "sl-cli.yaml"
		if _, err := os.Stat(filename); err == nil {
			fmt.Printf("⚠️  File '%s' already exists. Aborting to prevent overwrite.\n", filename)
			return
		}

		if err := os.WriteFile(filename, []byte(defaultConfigContent), 0o644); err != nil {
			fmt.Printf("❌ Failed to write file: %s\n", err)
			return
		}

		fmt.Printf("✅ Example config generated: %s\n", filename)
		fmt.Println("You can edit it and run 'sl-cli config check' to validate.")
	},
}

func init() {
	// 挂载子命令
	configCmd.AddCommand(checkCmd)
	configCmd.AddCommand(initCmd)

	// 挂载到 root
	rootCmd.AddCommand(configCmd)
}

// 嵌入的默认配置文件内容
const defaultConfigContent = `
# sl-cli Configuration File
# -------------------------
# This file defines the dynamic commands available in sl-cli.

commands:
  # ------------------------------------------------------------------
  # Example 1: HTTP Request (RESTful API)
  # ------------------------------------------------------------------
  - name: "myip"
    usage: "Get public IP address"
    type: "http"
    api:
      url: "https://httpbin.org/ip"
      method: "GET"
      # Optional: Add headers
      # headers:
      #   Authorization: "Bearer ${MY_TOKEN}"

  # ------------------------------------------------------------------
  # Example 2: HTTP with Pipe (JSON Processing)
  # ------------------------------------------------------------------
  - name: "weather"
    usage: "Get weather for a city (usage: sl-cli weather London)"
    type: "http"
    api:
      url: "https://goweather.herokuapp.com/weather/{{index .args 0}}"
      method: "GET"
      # Pipe the output to 'jq' for pretty printing
      pipe:
        command: "jq"
        args: ["."]

  # ------------------------------------------------------------------
  # Example 3: Shell Script
  # ------------------------------------------------------------------
  - name: "greet"
    usage: "Run a shell script with arguments"
    type: "shell"
    script: |
      echo "--------------------------------"
      echo "Hello, {{index .args 0}}!"
      echo "Current Dir: $(pwd)"
      echo "--------------------------------"

  # ------------------------------------------------------------------
  # Example 4: System Command Alias
  # ------------------------------------------------------------------
  - name: "ll"
    usage: "List files with details (alias for ls -laG)"
    type: "system"
    command: "ls"
    args: ["-l", "-a", "-G"]

  # ------------------------------------------------------------------
  # Example 5: Nested Commands
  # ------------------------------------------------------------------
  - name: "dev"
    usage: "Developer tools"
    subcommands:
      - name: "info"
        usage: "Show dev environment info"
        type: "shell"
        script: "go version && echo 'Node: ' $(node -v)"
`
