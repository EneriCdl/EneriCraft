//go:build !windows

package server

import "os/exec"

func setDetached(cmd *exec.Cmd) {
	// Linux/macOS: 默认行为即可，子进程独立于父进程
	cmd.SysProcAttr = nil
}
