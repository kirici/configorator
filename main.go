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
