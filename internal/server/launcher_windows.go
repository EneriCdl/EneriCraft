//go:build windows

package server

import (
	"os/exec"
	"syscall"
)

const (
	CREATE_NEW_PROCESS_GROUP    = 0x00000200
	CREATE_BREAKAWAY_FROM_JOB   = 0x01000000
	DETACHED_PROCESS            = 0x00000008
	CREATE_NO_WINDOW            = 0x08000000
)

func setDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NEW_PROCESS_GROUP | CREATE_BREAKAWAY_FROM_JOB,
		HideWindow:    true,
	}
}
