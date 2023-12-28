//go:build !windows

package platform

import (
	"log/slog"
	"os"
	"os/exec"
)

func Packer() (string, string) {
	const (
		url = "https://releases.hashicorp.com/packer/1.10.0/packer_1.10.0_linux_amd64.zip"
		app = "packer"
	)
	return url, app
}

func Terminate(cmd *exec.Cmd) {
	err := cmd.Process.Signal(os.Interrupt)
	if err != nil {
		slog.Error("Could not kill child", "err", err)
	}
}
