//go:build unix

package executor

import (
	"os/exec"
	"syscall"
)

// setProcessGroupKill 让命令运行在独立进程组中，取消时 kill 整个进程组
func setProcessGroupKill(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
