package main

import (
	"embed"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	// Channel will be used to propagate OS interrupts to child procs
	var c chan os.Signal = trapSIGTERM()
	go packer.Fetch()

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
	http.HandleFunc("/submit", submitHandler(submitTpl, c))

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
func submitHandler(tpl *template.Template, c chan os.Signal) http.HandlerFunc {
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
		err = packer.Exec(c)
		if err != nil {
			slog.Error("Packer failed to launch.", "err", err)
		}
	}
}

func trapSIGTERM() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		slog.Info("Received SIGTERM, exiting.")
		os.Exit(1)
	}()
	return c
}
