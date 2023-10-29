package main

import (
	"log"
	"net/http"
	"text/template"
)

type SearchResult struct {
    Title string
}

func main() {
    http.Handle("/assets/", http.FileServer(http.Dir(".")))

    http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
        // todo store in memory
        tmpl := template.Must(template.ParseFiles("./templates/index.html"))
        tmpl.Execute(w, nil)
    })

    http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request){
        query := r.URL.Query()
        str := query.Get("query")

        tmpl := template.Must(template.ParseFiles("./templates/search_results.html"))
        data := map[string][]SearchResult{
            "Results": {{"test1"}, {"test2"}, {"test3"}, {str}},
        }
        tmpl.Execute(w, data)
    })

    log.Println("server running")
    log.Fatal(http.ListenAndServe(":7323", nil))
}

