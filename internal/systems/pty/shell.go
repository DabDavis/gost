package pty

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// startShell launches an interactive shell session inside a PTY.
func startShell(shell string) (*os.File, *exec.Cmd, error) {
	cmd := exec.Command(shell, "-i")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, err
	}
	return f, cmd, nil
}

