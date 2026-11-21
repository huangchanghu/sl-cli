package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"sl-cli/internal/config"
)

// DefaultCompletionTimeout 是补全脚本的默认超时时间
const DefaultCompletionTimeout = 3 * time.Second

// ================= Completion Processor =================

// Complete 执行补全脚本，返回补全候选（每项可为 "value\tdescription" 格式）。
// 脚本通过 /bin/sh -c <script> sl-cli <args...> 执行，已输入的参数即 $1..$n，
// 当前正在输入的词通过 $SL_CLI_COMP_CUR 传入。
func Complete(cfg *config.CompletionConfig, cmdPath string, args []string, toComplete string, vars map[string]string) ([]string, error) {
	timeout, err := ParseCompletionTimeout(cfg.Timeout)
	if err != nil {
		return nil, err
	}

	// 模板主要用于引用 {{.vars.KEY}}；参数建议通过位置参数访问，避免半成品输入破坏脚本语法
	script, err := renderTemplate(cfg.Script, args, resolveVars(vars, args))
	if err != nil {
		return nil, fmt.Errorf("render completion script error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// $0 = sl-cli，$1..$n = 已输入的参数
	cmd := exec.CommandContext(ctx, "/bin/sh", append([]string{"-c", script, "sl-cli"}, args...)...)
	cmd.Env = append(os.Environ(),
		"SL_CLI_COMP_CUR="+toComplete,
		"SL_CLI_COMP_CMD="+cmdPath,
	)
	// 补全过程中不能交互，stdin/stderr 都不连接终端
	cmd.Stdin = nil
	cmd.Stderr = nil
	// 超时时连同管道中的子进程一起 kill，否则子进程持有 stdout 会让 Output() 一直阻塞
	setProcessGroupKill(cmd)
	cmd.WaitDelay = 500 * time.Millisecond

	out, err := cmd.Output()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("completion script timed out after %s", timeout)
	}
	if err != nil {
		return nil, fmt.Errorf("completion script failed: %w", err)
	}

	return parseCandidates(string(out), toComplete, cfg.Filter), nil
}

// ParseCompletionTimeout 解析 timeout 配置，为空时返回默认值
func ParseCompletionTimeout(s string) (time.Duration, error) {
	if s == "" {
		return DefaultCompletionTimeout, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid completion timeout '%s': %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("completion timeout must be positive, got '%s'", s)
	}
	return d, nil
}

// parseCandidates 按行拆分脚本输出，filter 为 prefix（默认）时只保留 value 部分以 toComplete 开头的候选
func parseCandidates(out, toComplete, filter string) []string {
	var candidates []string
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if filter != "none" {
			value, _, _ := strings.Cut(line, "\t")
			if !strings.HasPrefix(value, toComplete) {
				continue
			}
		}
		candidates = append(candidates, line)
	}
	return candidates
}
