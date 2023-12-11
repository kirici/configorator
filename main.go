package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/template"

	"github.com/kirici/configorator/model"
)

//go:embed templates/*
var content embed.FS

func main() {
	trapSIGTERM()
	profile := flatten(mapJSON("profile-config.json"))

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
		err := indexTemplate.Execute(w, profile)
		if err != nil {
			fmt.Printf("ERROR: Templates: %s", err)
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
		writeJSON(config, "profile-config-output.json")
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

func mapJSON(f string) map[string]interface{} {
	c, err := os.ReadFile(f)
	if err != nil {
		fmt.Printf("ERROR: Could not read file: %s", err)
	}
	result := make(map[string]interface{})
	json.Unmarshal([]byte(c), &result)
	return result
}

func flatten(m map[string]interface{}) map[string]interface{} {
	o := map[string]interface{}{}
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		case []interface{}:
			for i := 0; i < len(child); i++ {
				o[k+"."+strconv.Itoa(i)] = child[i]
			}
		default:
			o[k] = v
		}
	}
	return o
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
