package executor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"sl-cli/internal/config"

	"github.com/briandowns/spinner"
)

// Run 根据配置类型执行具体的逻辑
func Run(cfg config.CommandConfig, args []string) error {
	switch cfg.Type {
	case "http":
		return runHTTP(cfg, args)
	case "shell":
		return runShell(cfg, args)
	case "system":
		return runSystem(cfg, args)
	default:
		return fmt.Errorf("unknown command type: %s", cfg.Type)
	}
}

// ================= HTTP Processor =================

func runHTTP(cfg config.CommandConfig, args []string) error {
	// 1. 处理 URL 模板 (支持 {{.args.0}} 和 ${ENV})
	url, err := renderTemplate(cfg.API.URL, args)
	if err != nil {
		return fmt.Errorf("render url error: %w", err)
	}
	url = os.ExpandEnv(url)

	// 2. 处理 Body
	bodyStr := ""
	if cfg.API.Body != "" {
		bodyStr, err = renderTemplate(cfg.API.Body, args)
		if err != nil {
			return fmt.Errorf("render body error: %w", err)
		}
		bodyStr = os.ExpandEnv(bodyStr)
	}

	// 3. 创建 Request
	req, err := http.NewRequest(cfg.API.Method, url, strings.NewReader(bodyStr))
	if err != nil {
		return err
	}

	// 4. 处理 Headers (支持环境变量替换)
	for k, v := range cfg.API.Headers {
		expandedVal := os.ExpandEnv(v)
		req.Header.Set(k, expandedVal)
	}

	// 启动 Spinner ---
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond) // 14号是常用的点点点风格
	s.Suffix = fmt.Sprintf(" Requesting %s...", url)
	s.Color("cyan") // Mac 终端对 cyan 支持很好
	s.Start()

	// 5. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	s.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 只有状态码为 2xx 时才认为是“成功”，才执行管道命令
	// 否则直接输出错误信息或原始 Body，避免 jq 解析 HTML 报错
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("HTTP Request failed with status: %d %s\n", resp.StatusCode, resp.Status)
		// 依然输出 Body 以便调试错误信息
		_, _ = io.Copy(os.Stdout, resp.Body)
		return fmt.Errorf("http request failed")
	}

	// 处理管道逻辑
	if cfg.API.Pipe.Command != "" {
		// 准备管道命令参数（支持环境变量替换）
		pipeCmdName := cfg.API.Pipe.Command
		var pipeArgs []string
		for _, arg := range cfg.API.Pipe.Args {
			pipeArgs = append(pipeArgs, os.ExpandEnv(arg))
		}

		// 创建系统命令
		cmd := exec.Command(pipeCmdName, pipeArgs...)

		// 【关键】将 HTTP Response Body 直接接入命令的标准输入
		cmd.Stdin = resp.Body
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// 执行管道命令
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pipe execution failed: %w", err)
		}
		return nil
	}

	// 未配置管道命令，直接输出原始 Body
	_, err = io.Copy(os.Stdout, resp.Body)
	fmt.Println()
	return err
}

// ================= Shell Processor =================

func runShell(cfg config.CommandConfig, args []string) error {
	// 允许在脚本中使用模板参数，例如 echo {{.args.0}}
	scriptContent, err := renderTemplate(cfg.Script, args)
	if err != nil {
		return err
	}
	scriptContent = os.ExpandEnv(scriptContent)

	// 默认使用 sh -c 执行
	cmd := exec.Command("/bin/sh", "-c", scriptContent)

	// 绑定标准输入输出，支持交互
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ================= System Processor =================

func runSystem(cfg config.CommandConfig, args []string) error {
	// System 模式下，配置中的 Args 是基础参数，命令行输入的 args 追加在后面
	// 例如配置: git log; 输入: sl-cli git-log -n 5
	// 最终执行: git log -n 5

	finalArgs := append(cfg.Args, args...)

	cmd := exec.Command(cfg.Command, finalArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ================= Helper: Template Rendering =================

func renderTemplate(tplStr string, args []string) (string, error) {
	if tplStr == "" {
		return "", nil
	}

	// 准备模板数据
	data := map[string]interface{}{
		"args": args,
	}

	tmpl, err := template.New("cmd").Parse(tplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
