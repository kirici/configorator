//go:build windows

package platform

import (
	"log/slog"
	"os/exec"

	"github.com/iwdgo/sigintwindows"
)

func Packer() (string, string) {
	const (
		url = "https://releases.hashicorp.com/packer/1.10.0/packer_1.10.0_windows_amd64.zip"
		app = "packer.exe"
	)
	return url, app
}

func Terminate(cmd *exec.Cmd) {
	err := sigintwindows.SendCtrlBreak(cmd.Process.Pid)
	if err != nil {
		slog.Error("Could not kill child", "err", err)
	}
}
