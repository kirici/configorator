package main

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"text/template"

	"github.com/kirici/configorator/model"
	"github.com/pkg/browser"
)

//go:embed templates/*
var content embed.FS

func main() {
	trapSIGTERM()
	go fetch()

	// Parse templates during server startup
	indexTemplate, err := template.ParseFS(content, "templates/index.html", "templates/header.html")
	if err != nil {
		panic(err)
	}

	submitTemplate, err := template.ParseFS(content, "templates/submit.html", "templates/header.html")
	if err != nil {
		panic(err)
	}

	// Requests to "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := indexTemplate.Execute(w, nil)
		if err != nil {
			log.Fatalf("ERROR: Templates: %s", err)
		}
	})

	// POST "/submit"
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		err := submitTemplate.Execute(w, nil)
		if err != nil {
			log.Fatalf("ERROR: Template error: %s", err)
		}
		// Retrieve form data
		config := *model.ParseValues(r)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		writeJSON(config, "profile-config-output.json")
		go execPacker()
	})

	// Start the server
	port := "8080"
	fmt.Println("Starting server at", port)
	browser.OpenURL("http://127.0.0.1:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func trapSIGTERM() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Received SIGTERM, exiting.")
		os.Exit(1)
	}()
}

func writeJSON(input any, filename string) {
	jsonData, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		fmt.Printf("ERROR: Could not marshal JSON: %s", err)
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("ERROR: Could not create file: %s", err)
	}
	defer f.Close()
	f.Write(jsonData)
}

func fetch() {
	// url := link
	url := "https://releases.hashicorp.com/packer/1.10.0/packer_1.10.0_windows_amd64.zip"
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
	}
}

func execPacker() {
	cmd := exec.Command("./tools/packer.exe", "build", "-force", "-on-error=abort", "-only", "kx-main-virtualbox", "-var", "compute_engine_build=false", "-var", "memory=8192", "-var", "cpus=2", "-var", "video_memory=128", "-var", "hostname=kx-main", "-var", "domain=kx-as-code.local", "-var", "version=0.8.16", "-var", "kube_version=1.27.4-00", "-var", "vm_user=kx.hero", "-var", "vm_password=L3arnandshare", "-var", "git_source_url=https://github.com/Accenture/kx.as.code.git", "-var", "git_source_branch=main", "-var", "git_source_user=", "-var", "git_source_token=", "-var", "base_image_ssh_user=vagrant", "./kx-main-local-profiles.json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Could not attach stdout: %s", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Could not attach stderr: %s", err)
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("Could not start command: %s", err) // Let user correct the circumstances and retry
	}
	// Start() does not wait for the command to complete, this ensures that main doesnt exit before cmd does
	defer cmd.Wait()
	childLog, err := os.OpenFile("configo.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o666)
	if err != nil {
		log.Printf("Could not open log file: %s", err)
	}
	// copy packer output to terminal and log file
	go io.Copy(io.MultiWriter(os.Stdout, childLog), stdout)
	go io.Copy(io.MultiWriter(os.Stderr, childLog), stderr)
}

// C:\projects\kx.as.code\base-vm\build\jenkins\jenkins_home\tools\biz.neustar.jenkins.plugins.packer.PackerInstallation\packer-windows\packer.exe build -force -on-error=abort -only kx-main-virtualbox -var compute_engine_build=false -var memory=8192 -var cpus=2 -var video_memory=128 -var hostname=kx-main -var domain=kx-as-code.local -var version=0.8.16 -var kube_version=1.27.4-00 -var vm_user=kx.hero -var vm_password=L3arnandshare -var git_source_url=https://github.com/Accenture/kx.as.code.git -var git_source_branch=main -var git_source_user= -var git_source_token= -var base_image_ssh_user=vagrant ./kx-main-local-profiles.json
// unwrap>
// packer.exe build -force -on-error=abort -only kx-main-virtualbox \
// -var compute_engine_build=false \
// -var memory=8192 \
// -var cpus=2 \
// -var video_memory=128 \
// -var hostname=kx-main \
// -var domain=kx-as-code.local \
// -var version=0.8.16 \
// -var kube_version=1.27.4-00 \
// -var vm_user=kx.hero \
// -var vm_password=L3arnandshare \
// -var git_source_url=https://github.com/Accenture/kx.as.code.git \
// -var git_source_branch=main \
// -var git_source_user= \
// -var git_source_token= \
// -var base_image_ssh_user=vagrant \
// ./kx-main-local-profiles.json
