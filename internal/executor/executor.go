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

	"sl-cli/internal/config"
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

	// 5. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 6. 输出结果到终端
	// 这里直接输出原始 Body，未来可以优化为 JSON Pretty Print
	_, err = io.Copy(os.Stdout, resp.Body)
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
