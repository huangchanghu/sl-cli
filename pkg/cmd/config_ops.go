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

		// 3. 递归逻辑校验
		errCount := 0
		for i, c := range cfg.Commands {
			// 顶层命令路径直接用名字，如果没有名字则用索引
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
		// 如果名字都没有，path 也是构造出来的，继续校验意义不大，但在递归中最好继续检查子项
	}

	// 2. 结构校验：必须是 "有效的功能命令" 或者 "包含子命令的组"
	// 如果没有 Type 且没有 SubCommands，那就是个空壳
	if c.Type == "" && len(c.SubCommands) == 0 {
		fmt.Printf("❌ Error in [%s]: Must specify 'type' (http/shell/system) OR have 'subcommands'.\n", path)
		errs++
	}

	// 3. 类型校验 (如果指定了 Type)
	if c.Type != "" {
		validTypes := map[string]bool{"http": true, "shell": true, "system": true}
		if !validTypes[c.Type] {
			fmt.Printf("❌ Error in [%s]: Invalid type '%s'. Must be http, shell, or system.\n", path, c.Type)
			errs++
		}

		// 4. 字段校验 (根据 Type)
		switch c.Type {
		case "http":
			if c.API.URL == "" {
				fmt.Printf("❌ Error in [%s]: Type is http but 'api.url' is missing.\n", path)
				errs++
			}
			// 校验 Pipes
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

	// 5. 递归校验子命令
	for _, sub := range c.SubCommands {
		subPath := path + " -> " + sub.Name
		if sub.Name == "" {
			subPath = path + " -> [Unnamed]"
		}
		errs += validateCommand(sub, subPath)
	}

	return errs
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

		// 写入文件时使用 0644 权限
		if err := os.WriteFile(filename, []byte(defaultConfigContent), 0o644); err != nil {
			fmt.Printf("❌ Failed to write file: %s\n", err)
			return
		}

		fmt.Printf("✅ Example config generated: %s\n", filename)
		fmt.Println("You can edit it and run 'sl-cli config check' to validate.")
	},
}

func init() {
	configCmd.AddCommand(checkCmd)
	configCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
}

// 嵌入的默认配置文件内容 (已修正 pipes 格式)
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
  # Example 2: HTTP with Pipes (Chained Processing)
  # ------------------------------------------------------------------
  - name: "weather"
    usage: "Get weather for a city (usage: sl-cli weather London)"
    type: "http"
    api:
      url: "https://goweather.herokuapp.com/weather/{{index .args 0}}"
      method: "GET"
      # Pipeline: API Response -> jq -> grep -> Stdout
      pipes:
        - command: "jq"
          args: ["."]
        - command: "grep"
          args: ["temperature"]

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
