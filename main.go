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
		configType := evalConfig(r)
		writeJSON(configType, "profile-config.json")
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
	f.Close()
}

func tmplParse() *template.Template {
	tmpl, err := template.New("base.html").Funcs(template.FuncMap{
		"IsString": checkString,
		"IsInt":    checkInt,
		"IsBool":   checkBool,
	}).ParseFiles("base.html")
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		panic(err)
	}
	return tmpl
}

// TODO: implementation
// func genHTML(t reflect.Kind) {
// 	var tmpl, key string
// 	switch t {
// 	case reflect.Bool:
// 		tmpl = `<label>` + key + `:</label>
// 				<input type="checkbox" id="` + key + `" name="` + key + `" /><br />`
// 	case reflect.String:
// 		tmpl = `<label>` + key + `:</label>
// 				<input type="text" id="` + key + `" name="` + key + `" /><br />`
// 	case reflect.Int:
// 		tmpl = `<label>` + key + `:</label>
// 				<input type="number" id="` + key + `" name="` + key + `" /><br />`
// 	default:
// 		return
// 	}
// 	var result bytes.Buffer
// 	return
// }

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

// Likely to be deprecated together with FuncMap if custom generator instead of html/template
func checkString(i interface{}) bool {
	_, ok := i.(string)
	return ok
}

func checkInt(i interface{}) bool {
	_, ok := i.(int)
	return ok
}

func checkBool(i interface{}) bool {
	_, ok := i.(bool)
	return ok
}
