//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// npx runs the server as its own child, so the server gets a process
// group and signals go to the whole group.
func setGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func signalGroup(cmd *exec.Cmd, s os.Signal) {
	if ss, ok := s.(syscall.Signal); ok {
		if syscall.Kill(-cmd.Process.Pid, ss) == nil {
			return
		}
	}
	_ = cmd.Process.Signal(s)
}
