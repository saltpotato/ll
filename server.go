package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	tt "text/template"
	"time"
)

// Position a position in a company, a cv-line
type Position struct {
	From, Type, Org, OrgLink, Where, Does, MyArea string
	Topics                                        []string
	Href                                          []string
}

//CV struct
type CV struct {
	Titles       map[string]string
	About        map[string]string
	Professional []Position `json:"professional"`
	Study        []Position `json:"study"`
	Education    []Position `json:"education"`
	Experience   map[string]string
	Interests    string
}

var funcMap = map[string]interface{}{
	"Split": strings.Split,
	"GetHREFText": func(position Position, token string) string {
		if strings.HasPrefix(token, "\\") {

			numPrefixed := token[1:]
			numNTask := strings.Split(numPrefixed, ",")

			if len(numNTask) < 2 {
				return ""
			}

			idx, _ := strconv.Atoi(numNTask[0])
			return position.Href[idx]
		}
		return ""
	},
	"StripHREFIndicator": func(token string) string {

		idx := strings.Index(token, ",")
		if idx != -1 && idx+1 < len(token) {
			return token[idx+1:]
		}
		return ""
	},
	"StartsWithDash": func(task string) bool {
		return strings.HasPrefix(task, "-")
	},
	"enDe": func(en, de string) string {
		isEn := cvJSONFile == "positions_en.json"
		if isEn {
			return en
		}
		return de
	},
	"lang": func() string {
		isEn := cvJSONFile == "positions_en.json"
		if isEn {
			return "en"
		}
		return "de"
	},
}

var cvJSONFile = "positions_de.json"

//ServeIndexHTML serve index html
func ServeIndexHTML(w http.ResponseWriter, req *http.Request) {

	jsonFile, err := os.Open(cvJSONFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var cv CV

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &cv)

	tmpl, err := template.New("index_template.html").Funcs(funcMap).ParseFiles("pages/index_template.html")
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "html")

	err = tmpl.Execute(w, cv)
	if err != nil {
		fmt.Println(err)
	}
}

func jsonFileFromUrl(urlPath string) string {
	if strings.Contains(urlPath, "/de") {
		return "positions_de.json"
	} else {
		return "positions_en.json"
	}
}

func switchToEN(w http.ResponseWriter, r *http.Request) {
	cvJSONFile = "positions_en.json"
	ServeIndexHTML(w, r)
}
func switchToDE(w http.ResponseWriter, r *http.Request) {
	cvJSONFile = "positions_de.json"
	ServeIndexHTML(w, r)
}

func exportLatex(w http.ResponseWriter, r *http.Request) {

	jsonFile, err := os.Open(jsonFileFromUrl(r.URL.Path))
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var cv CV

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &cv)

	tmpl, err := tt.New("cv.tt").Funcs(funcMap).ParseFiles("texttemplate/cv.tt")
	if err != nil {
		fmt.Println(err)
		return
	}

	textOutPath := "wwwroot/tex/"
	texOutFilePath := textOutPath + "cv.tex"
	postex, err := os.Create(texOutFilePath)
	defer postex.Close()

	err = tmpl.Execute(postex, cv)
	if err != nil {
		fmt.Println(err)
	}

	cmd := exec.Command("pdflatex", "cv.tex")
	fmt.Println("launching pdflatex")
	cmd.Dir = textOutPath

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Command finished with error: %v \n", err)
		log.Printf(string(out))
	}

	http.ServeFile(w, r, strings.Replace(texOutFilePath, ".tex", ".pdf", 1))
}

func toWwwRoot(w http.ResponseWriter, r *http.Request) {
	path := "wwwroot/" + r.URL.Path[1:]
	if strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".jpg") {
		http.ServeFile(w, r, path)
	} else {

		file, err := os.Open(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		http.ServeContent(w, r, "x", time.Time{}, file)
	}
}

func main() {

	http.HandleFunc("/en/", switchToEN)
	http.HandleFunc("/de/", switchToDE)
	http.HandleFunc("/en/cv.pdf", exportLatex)
	http.HandleFunc("/de/cv.pdf", exportLatex)
	http.HandleFunc("/node_modules/", toWwwRoot)
	http.HandleFunc("/images/", toWwwRoot)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		requestPath := r.URL.Path[1:]
		if len(requestPath) == 0 {
			ServeIndexHTML(w, r)
		} else {
			path := "/pages/" + requestPath
			http.ServeFile(w, r, path)
		}

	})

	err := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
