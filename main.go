package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
)

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
    http.HandleFunc("/search", searchHandler)

    log.Printf("server running on port %d\n", *port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func rootHander(w http.ResponseWriter, _ *http.Request) {
    // todo store in memory
    tmpl := template.Must(template.ParseFiles("./templates/index.html"))
    tmpl.Execute(w, nil)
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

