package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
)

type SuggestionResult struct {
    Suggestion string
    // maybe add from history vs no history
}

type SearchResult struct {
    Title string
}

func main() {
    port := flag.Int("port", 7323, "port")
    knowledge_base_path := flag.String("path", "", "where are txt documents")
    flag.Parse()

    if *knowledge_base_path == "" {
        log.Fatalln("path must be provided")
    }


    http.Handle("/assets/", http.FileServer(http.Dir(".")))
    http.HandleFunc("/", rootHander)
    http.HandleFunc("/suggest", suggestHandler)
    http.HandleFunc("/search", searchHandler)

    log.Printf("server running on port %d\n", *port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func rootHander(w http.ResponseWriter, _ *http.Request) {
    tmpl := template.Must(template.ParseFiles("./templates/index.html"))
    tmpl.Execute(w, nil)
}

func suggestHandler(w http.ResponseWriter, r *http.Request) {
    query := strings.TrimSpace(r.URL.Query().Get("query"))
    if query == "" {
        return
    }

    tmpl := template.Must(template.ParseFiles("./templates/suggestion_results.html"))
    data := map[string][]SuggestionResult{
        "Suggestions": {{"test1"}, {"test2"}, {"test3"}, {query}},
    }
    log.Println("ret")
    tmpl.Execute(w, data)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    str := strings.TrimSpace(query.Get("query"))

    if str == "" {
        w.Write([]byte(""))
        return
    }

    tmpl := template.Must(template.ParseFiles("./templates/search_results.html"))
    data := map[string][]SearchResult{
        "Results": {{"test1"}, {"test2"}, {"test3"}, {str}},
    }
    tmpl.Execute(w, data)
}
