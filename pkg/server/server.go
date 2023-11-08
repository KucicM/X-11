package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type ServerCfg struct {
    TemplatesPath string `json:"templates-path"`
    AssetsPath string `json:"assets-path"`
    ResultsPerPage int `json:"results-per-page"`
    Port int `json:"port"`
    SiCfg SearchIndexCfg `json:"search-index"`
}

type server struct {
    resultsPerPage int
    searchIndex *SearchIndex

    // to be replaced with in memory templates
    autocompleteTemplatePath string
}

func StartServer(cfg ServerCfg) {

    srv := &server{
        autocompleteTemplatePath: fmt.Sprintf("%s/autocomplete_results.html", cfg.TemplatesPath),
        searchIndex: NewSearchIndex(cfg.SiCfg),
        resultsPerPage: cfg.ResultsPerPage,
    }

    assetsPath, _ := strings.CutSuffix(cfg.AssetsPath, "/")
    assetsPath, _ = strings.CutSuffix(assetsPath, "assets")
    http.Handle("/assets/", http.FileServer(http.Dir(assetsPath)))
    http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
        indexPage := fmt.Sprintf("%s/assets/index.html", assetsPath)
        file, err := os.ReadFile(indexPage)
        if err != nil {
            log.Printf("ERROR: failed to read index.html, %s", err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.Write(file)
    })

    http.HandleFunc("/autocomplete", srv.autocompleteHandler)
    http.HandleFunc("/search", srv.searchHandler)

    err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
    srv.stop()
    log.Fatal(err)
}

func (s *server) stop() {
    log.Println("stopping server...")
}


func (s *server) autocompleteHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    //w.Header().Add("Cache-Control", "private, max-age=3600")
    query := strings.TrimSpace(r.URL.Query().Get("query"))
    if query == "" {
        return
    }

    tmpl := template.Must(template.ParseFiles(s.autocompleteTemplatePath))

    // todo fetch actula data

    data := []string{"test 1", "test 2"}
    tmpl.Execute(w, data)
}

func (s *server) searchHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    query := strings.TrimSpace(r.URL.Query().Get("query"))

    res, err := s.searchIndex.Search(query, s.resultsPerPage)
    if err != nil {
        log.Printf("ERROR: /serach error %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    tmpl := template.Must(template.ParseFiles("./templates/search_results.html"))
    tmpl.Execute(w, res)
}
