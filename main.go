package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	"github.com/kirici/configorator/model"
)

//go:embed templates/*
var content embed.FS

func main() {
	trapSIGTERM()

	// Parse templates during server startup
	indexTemplate, err := template.ParseFS(content, "templates/index.html")
	if err != nil {
		panic(err)
	}

	submitTemplate, err := template.ParseFS(content, "templates/submit.html")
	if err != nil {
		panic(err)
	}

	// Requests to "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pageFields := model.Profile{} // TODO: Range over Profile to template fields
		err := indexTemplate.Execute(w, pageFields)
		if err != nil {
			fmt.Printf("ERROR: Template error: %s", err)
			return
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
			fmt.Printf("ERROR: Template error: %s", err)
			return
		}
		// Retrieve form data
		config := *model.ParseValues(r)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		writeJSON(config, "profile-config.json")
	})

	// Start the server
	port := "8080"
	fmt.Println("Starting server at", port)
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
