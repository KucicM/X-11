package server

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net/http"
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

    http.HandleFunc("/", srv.rootHandler)
    http.HandleFunc("/autocomplete", srv.autocompleteHandler)
    http.HandleFunc("/search", srv.searchHandler)
    http.HandleFunc("/articleClick", srv.articleClickHandler)

    err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
    srv.stop()
    log.Fatal(err)
}

func (s *server) stop() {
    log.Println("stopping server...")
}

func (s *server) rootHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{"Query": "", "SearchResults":""}
    buf, err := s.getFullPageRender(data)
    if err != nil {
        log.Printf("ERROR: / full page render %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    if _, err := buf.WriteTo(w); err != nil {
        log.Printf("ERROR: / writing response %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
}

func (s *server) autocompleteHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    query := strings.TrimSpace(r.URL.Query().Get("query"))
    if query == "" {
        return
    }

    tmpl := template.Must(template.ParseFiles(s.autocompleteTemplatePath))

    // todo fetch actula data
    var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    data := make([]string, 5)
    for i := range data {

        word := []rune(query)
        for j := 0; j < rand.Intn(5); j++ {
            word = append(word, letterRunes[j])
        }
        data[i] = string(word)
    }
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

    tmpl, err := template.ParseFiles("./templates/search_results.html")
    if err != nil {
        log.Printf("ERROR: /search parsing template %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    var resBuf bytes.Buffer
    if err := tmpl.Execute(&resBuf, res); err != nil {
        log.Printf("ERROR: /search executing template %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Header().Set("hx-push-url", fmt.Sprintf("/search?query=%s", query))
    w.Header().Set("hx-history-restore", "true")

    if isPartialRequest(r.Header) {
        if _, err := resBuf.WriteTo(w); err != nil {
            log.Printf("ERROR: /search writing response %s", err)
            w.WriteHeader(http.StatusInternalServerError)
            return 
        }
    } else {
        buf, err := s.getFullPageRender(map[string]string{"Query": query, "SearchResults": resBuf.String()})
        if err != nil {
            log.Printf("ERROR: /serach full page response %s", err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        if _, err := buf.WriteTo(w); err != nil {
            log.Printf("ERROR: /serach writing response %s", err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
    }
}

func (s *server) getFullPageRender(data map[string]string) (bytes.Buffer, error) {
    var res bytes.Buffer
    tmpl, err := template.ParseFiles("./templates/index.html")
    if err != nil {
        return res, err
    }
    if err := tmpl.Execute(&res, data); err != nil {
        return res, err
    }
    return res, nil
}

func (s *server) articleClickHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    w.Header().Add("HX-Redirect", "https://www.example.com")
    w.WriteHeader(http.StatusOK)
}

func isPartialRequest(header http.Header) bool {
    return header.Get("Hx-Request") == "true"
}
