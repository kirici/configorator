package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/kirici/configorator/model"
)

func main() {
	port := ":8080"
	http.HandleFunc("/", handleReq)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fmt.Println("Starting server at", port)
	http.ListenAndServe(port, nil)
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		setConfigTypes := evalConfig(r)
		writeJSON(setConfigTypes, "profile-config.json")
		execPipeline("cat", "profile-config.json")
		http.ServeFile(w, r, "submit.html")
		return
	}
	tmpl := tmplParse()
	tmpl.Execute(w, nil)
}

// TODO: Try and avoid reflection using maps and interfaces
func evalConfig(r *http.Request) model.Profile {
	profile := model.Profile{}
	profileValue := reflect.ValueOf(&profile).Elem()
	for i := 0; i < profileValue.NumField(); i++ {
		field := profileValue.Field(i)
		fieldType := field.Type()
		if fieldType.Kind() == reflect.Struct {
			for j := 0; j < field.NumField(); j++ {
				subField := field.Field(j)
				subFieldType := subField.Type()
				subFieldName := field.Type().Field(j).Name
				// TODO: validate parsed values match subFieldType
				switch subFieldType.Kind() {
				case reflect.String:
					subFieldValue := r.FormValue(subFieldName)
					subField.SetString(subFieldValue)
				case reflect.Int64:
					subFieldValue, _ := strconv.ParseInt(r.FormValue(subFieldName), 10, 64)
					subField.SetInt(subFieldValue)
				case reflect.Bool:
					exists := r.Form.Has(firstToLower(subFieldName))
					subField.SetBool(exists)
				}
			}
		}
	}
	return profile
}

func writeJSON(profile model.Profile, filename string) {
	jsonData, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}
	defer f.Close()

	f.Write(jsonData)
}

func tmplParse() *template.Template {
	tmpl, err := template.New("base.html").ParseFiles("base.html")
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		panic(err)
	}
	return tmpl
}

func execPipeline(launcher string, config string) {
	out, err := exec.Command(launcher, config).Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", out)
}

func firstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}
