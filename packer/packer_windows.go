package packer

import (
	"archive/zip"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func simpleCheck(err error) {
	if err != nil {
		slog.Error("ERROR: %s", err)
		return
	}
}

func Fetch() {
	slog.Info("Downloading Packer.")
	const (
		url  = "https://releases.hashicorp.com/packer/1.10.0/packer_1.10.0_windows_amd64.zip"
		file = "packer.zip"
	)
	resp, err := http.Get(url)
	simpleCheck(err)
	defer resp.Body.Close()

	out, err := os.Create(file)
	simpleCheck(err)
	defer out.Close()

	io.Copy(out, resp.Body)
	unzip(file, "tools")
}

func unzip(file string, out string) {
	slog.Info("Unpacking Packer.")
	reader, err := zip.OpenReader(file)
	simpleCheck(err)
	defer reader.Close()

	for _, file := range reader.File {
		src, err := file.Open()
		simpleCheck(err)
		defer src.Close()

		relname := path.Join(out, file.Name)
		dir := path.Dir(relname)
		os.MkdirAll(dir, 0o777)
		dst, err := os.Create(relname)
		simpleCheck(err)
		defer dst.Close()

		io.Copy(dst, src)
	}
}

func Exec(sig chan os.Signal) error {
	cmd := exec.Command("./tools/packer.exe", "build", "-force", "-on-error=abort", "-only", "kx-main-virtualbox", "-var", "compute_engine_build=false", "-var", "memory=8192", "-var", "cpus=2", "-var", "video_memory=128", "-var", "hostname=kx-main", "-var", "domain=kx-as-code.local", "-var", "version=0.8.16", "-var", "kube_version=1.27.4-00", "-var", "vm_user=kx.hero", "-var", "vm_password=L3arnandshare", "-var", "git_source_url=https://github.com/Accenture/kx.as.code.git", "-var", "git_source_branch=main", "-var", "git_source_user=", "-var", "git_source_token=", "-var", "base_image_ssh_user=vagrant", "./kx-main-local-profiles.json")
	stdout, err := cmd.StdoutPipe()
	simpleCheck(err)
	stderr, err := cmd.StderrPipe()
	simpleCheck(err)
	err = cmd.Start()
	if err != nil {
		return errors.New(err.Error())
	}
	// Start() does not wait for the command to complete, this ensures that main doesnt exit before cmd does
	defer cmd.Wait()
	childLog, err := os.OpenFile("packer.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o666)
	simpleCheck(err)
	// copy packer output to terminal and log file
	go io.Copy(io.MultiWriter(os.Stdout, childLog), stdout)
	go io.Copy(io.MultiWriter(os.Stderr, childLog), stderr)
	go func() {
		<-sig
		cmd.Process.Signal(os.Interrupt)
	}()
	return nil
}
