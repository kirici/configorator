package packer

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func Fetch() {
	url := "https://releases.hashicorp.com/packer/1.10.0/packer_1.10.0_linux_amd64.zip"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("ERROR: Could not download packer; %s", err)
		return
	}
	defer resp.Body.Close()
	f := "packer.zip"
	out, err := os.Create(f)
	if err != nil {
		log.Printf("ERROR: Could not create file; %s", err)
		return
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	unzip("tools")
}

func unzip(name string) {
	f := "packer.zip"
	reader, err := zip.OpenReader(f)
	if err != nil {
		log.Printf("ERROR: Could not open zip file; %s", err)
		return
	}
	defer reader.Close()
	for _, file := range reader.File {
		in, _ := file.Open()
		defer in.Close()
		relname := path.Join(name, file.Name)
		dir := path.Dir(relname)
		os.MkdirAll(dir, 0o777)
		out, _ := os.Create(relname)
		defer out.Close()
		io.Copy(out, in)
		err = os.Chmod(out.Name(), 0o755)
		if err != nil {
			log.Printf("ERROR: Could not make file executable: %s", err)
		}
	}
}

func Exec(sig chan os.Signal) {
	cmd := exec.Command("./tools/packer", "build", "-force", "-on-error=abort", "-only", "kx-main-virtualbox", "-var", "compute_engine_build=false", "-var", "memory=8192", "-var", "cpus=2", "-var", "video_memory=128", "-var", "hostname=kx-main", "-var", "domain=kx-as-code.local", "-var", "version=0.8.16", "-var", "kube_version=1.27.4-00", "-var", "vm_user=kx.hero", "-var", "vm_password=L3arnandshare", "-var", "git_source_url=https://github.com/Accenture/kx.as.code.git", "-var", "git_source_branch=main", "-var", "git_source_user=", "-var", "git_source_token=", "-var", "base_image_ssh_user=vagrant", "./kx-main-local-profiles.json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("ERROR: Could not attach stdout: %s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("ERROR: Could not attach stderr: %s", err)
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("ERROR: Could not start command: %s", err) // Let user correct the circumstances and retry
	}
	// Start() does not wait for the command to complete, this ensures that main doesnt exit before cmd does
	defer cmd.Wait()
	childLog, err := os.OpenFile("configo.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o666)
	if err != nil {
		log.Printf("ERROR: Could not open log file: %s", err)
	}
	// copy packer output to terminal and log file
	go io.Copy(io.MultiWriter(os.Stdout, childLog), stdout)
	go io.Copy(io.MultiWriter(os.Stderr, childLog), stderr)
	go func() {
		<-sig
		cmd.Process.Signal(os.Interrupt)
	}()
}
