package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	"github.com/kirici/configorator/packer"
	"github.com/pkg/browser"
)

//go:embed templates/*
var content embed.FS

func main() {
	// Channel will be used to propagate OS interrupts to child procs
	var c chan os.Signal = trapSIGTERM()
	go packer.Fetch()

	// Parse templates during server startup
	indexTpl, err := template.ParseFS(content, "templates/index.html", "templates/header.html")
	if err != nil {
		log.Fatalf("Could not parse template: %s", err)
	}

	submitTpl, err := template.ParseFS(content, "templates/submit.html", "templates/header.html")
	if err != nil {
		log.Fatalf("Could not parse template: %s", err)
	}

	// Requests to "/"
	http.HandleFunc("/", serveIndex(indexTpl))
	// POST "/submit"
	http.HandleFunc("/submit", submitHandler(submitTpl, c))

	// Start the server
	port := "8080"
	fmt.Println("Starting server at", port)
	browser.OpenURL("http://127.0.0.1:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Indirection of returning an http Handler to enable passing parameters
func serveIndex(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tpl.Execute(w, nil)
		if err != nil {
			log.Fatalf("ERROR: Template error: %s", err)
		}
	}
}

// submitHandler validates the request method and passes the sigterm channel parameter on to the child proc
func submitHandler(tpl *template.Template, c chan os.Signal) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		err := tpl.Execute(w, nil)
		if err != nil {
			log.Fatalf("ERROR: Template error: %s", err)
		}
		go packer.Exec(c)
	}
}

func trapSIGTERM() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Received SIGTERM, exiting.")
		os.Exit(1)
	}()
	return c
}

// writeJSON is slated for removal
// func writeJSON(input any, filename string) {
// 	jsonData, err := json.MarshalIndent(input, "", "  ")
// 	if err != nil {
// 		fmt.Printf("ERROR: Could not marshal JSON: %s", err)
// 	}
// 	f, err := os.Create(filename)
// 	if err != nil {
// 		fmt.Printf("ERROR: Could not create file: %s", err)
// 	}
// 	defer f.Close()
// 	f.Write(jsonData)
// }

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
