package main

import (
	"log"
	"net/http"
	"text/template"
)

func main() {
    http.Handle("/assets/", http.FileServer(http.Dir(".")))

    http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
        // todo store in memory
        tmpl := template.Must(template.ParseFiles("./templates/index.html"))
        tmpl.Execute(w, nil)
    })

    http.HandleFunc("/search", func(_ http.ResponseWriter, r *http.Request){
        query := r.URL.Query()
        str := query.Get("query")
        log.Printf("got %+s\n", str)
    })

    log.Println("server running")
    log.Fatal(http.ListenAndServe(":7323", nil))
}

