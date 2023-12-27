package main

import (
	"archive/zip"
	"embed"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"text/template"

	"github.com/cli/browser"
	"github.com/kirici/configorator/packer"
)

//go:embed templates/*
var content embed.FS

func main() {
	f, err := os.OpenFile("configo.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o666)
	if err != nil {
		slog.Error("Could not open log file", "err", err)
	}
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stdout, f), nil))
	slog.SetDefault(logger)

	// Could be replaced by runtime checks
	url, bin := packer.PlatformVars()
	go fetch(url)

	// Parse templates during server startup
	indexTpl, err := template.ParseFS(content, "templates/index.html", "templates/header.html")
	if err != nil {
		slog.Error("Could not parse template", "err", err)
		os.Exit(1)
	}
	submitTpl, err := template.ParseFS(content, "templates/submit.html", "templates/header.html")
	if err != nil {
		slog.Error("Could not parse template", "err", err)
		os.Exit(1)
	}

	// Requests to "/"
	http.HandleFunc("/", serveIndex(indexTpl))
	// POST "/submit"
	http.HandleFunc("/submit", submitHandler(submitTpl, bin))

	// Start the server
	port := "8080"
	slog.Info("Starting server at http://127.0.0.1:" + port)
	browser.Stdout, browser.Stderr = io.Discard, io.Discard
	err = browser.OpenURL("http://127.0.0.1:" + port)
	if err != nil {
		slog.Error("Could not open browser", "err", err)
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("Could not start server", "err", err)
		os.Exit(1)
	}
}

// Indirection of returning an http Handler to enable passing parameters
func serveIndex(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := tpl.Execute(w, nil)
		if err != nil {
			slog.Error("Templating error", "err", err)
		}
	}
}

// submitHandler validates the request method and passes the sigterm channel parameter on to the child proc
func submitHandler(tpl *template.Template, bin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		err := tpl.Execute(w, nil)
		if err != nil {
			slog.Error("Templating error", "err", err)
		}
		slog.Info("Launching Packer.")
		err = execPacker(bin)
		if err != nil {
			slog.Error("Packer failed to launch.", "err", err)
		}
	}
}

func fetch(url string) {
	slog.Info("Downloading Packer.")
	const file = "packer.zip"

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

		err = os.Chmod(dst.Name(), 0o755)
		simpleCheck(err)
	}
}

func execPacker(bin string) error {
	sig := make(chan os.Signal, 2) // buffer should be equal to number of signals that can be sent
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	cmd := exec.Command("./tools/"+bin, "build", "-force", "-on-error=abort", "-only", "kx-main-virtualbox", "-var", "compute_engine_build=false", "-var", "memory=8192", "-var", "cpus=2", "-var", "video_memory=128", "-var", "hostname=kx-main", "-var", "domain=kx-as-code.local", "-var", "version=0.8.16", "-var", "kube_version=1.27.4-00", "-var", "vm_user=kx.hero", "-var", "vm_password=L3arnandshare", "-var", "git_source_url=https://github.com/Accenture/kx.as.code.git", "-var", "git_source_branch=main", "-var", "git_source_user=", "-var", "git_source_token=", "-var", "base_image_ssh_user=vagrant", "./kx-main-local-profiles.json")
	stdout, err := cmd.StdoutPipe()
	simpleCheck(err)
	stderr, err := cmd.StderrPipe()
	simpleCheck(err)
	err = cmd.Start()
	if err != nil {
		return errors.New(err.Error())
	}
	defer cmd.Start()
	go func() {
		<-sig
		slog.Info("Terminating Packer")
		err := cmd.Process.Kill()
		if err != nil {
			slog.Error("Could not kill child", "err", err)
		}
		os.Exit(127)
	}()
	childLog, err := os.OpenFile("packer.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o666)
	simpleCheck(err)
	// copy packer output to terminal and log file
	go io.Copy(io.MultiWriter(os.Stdout, childLog), stdout)
	go io.Copy(io.MultiWriter(os.Stderr, childLog), stderr)
	return nil
}

func simpleCheck(err error) {
	if err != nil {
		slog.Error("OS error", "err", err)
		return
	}
}
