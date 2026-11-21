//go:build !unix

package executor

import "os/exec"

// setProcessGroupKill 在非 unix 平台上退化为默认行为（只 kill 直接子进程，依赖 WaitDelay 兜底）
func setProcessGroupKill(cmd *exec.Cmd) {}
