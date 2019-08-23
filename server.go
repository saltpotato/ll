package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Position a position in a company, a cv-line
type Position struct {
	From, Type, Org, OrgLink, Where, Does, MyArea string
	Topics                                        []string
	Href                                          []string
}

//Personal Stammdaten
type Personal struct {
	BirthDate, BirthPlace, Address, Cell string
}

//Positions cv struct
type Positions struct {
	personal Personal
	Poss     []Position `json:"positions"`
}

//ServeIndexHTML serve index html
func ServeIndexHTML(w http.ResponseWriter, req *http.Request) {

	funcMap := template.FuncMap{
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
	}

	jsonFile, err := os.Open("positions.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Successfully Opened positions.json")

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var positions Positions

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &positions)

	tmpl, err := template.New("index_template.html").Funcs(funcMap).ParseFiles("pages/index_template.html")
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "html")

	err = tmpl.Execute(w, positions.Poss)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	http.HandleFunc("/bower_components/", func(w http.ResponseWriter, r *http.Request) {
		path := "wwwroot/" + r.URL.Path[1:]
		if strings.HasSuffix(path, ".css") {
			http.ServeFile(w, r, path)
		} else {

			file, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			http.ServeContent(w, r, "x", time.Time{}, file)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		requestPath := r.URL.Path[1:]
		if len(requestPath) == 0 {
			ServeIndexHTML(w, r)
		} else {
			path := "/pages/" + requestPath
			http.ServeFile(w, r, path)
		}

	})

	err := http.ListenAndServeTLS(":443", "https-server.crt", "https-server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
