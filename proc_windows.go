//go:build windows

package main

import (
	"os"
	"os/exec"
)

func setGroup(*exec.Cmd) {}

// Windows has no SIGTERM delivery; the server exits when stdin closes.
func signalGroup(cmd *exec.Cmd, _ os.Signal) { _ = cmd.Process.Kill() }
