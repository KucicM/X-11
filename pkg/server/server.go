package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type ServerCfg struct {
    TemplatesPath string `json:"templates-path"`
    AssetsPath string `json:"assets-path"`
    ResultsPerPage int `json:"results-per-page"`
    Port int `json:"port"`
    FtsCfg FullTextSearchCfg `json:"full-text-search"`
}

type server struct {
    resultsPerPage int
    fts *FullTextSearch

    // to be replaced with in memory templates
    autocompleteTemplatePath string
}

func StartServer(cfg ServerCfg) {
    start := time.Now()
    srv := &server{
        autocompleteTemplatePath: fmt.Sprintf("%s/autocomplete_results.html", cfg.TemplatesPath),
        fts: NewFullTextSearch(cfg.FtsCfg),
        resultsPerPage: cfg.ResultsPerPage,
    }

    assetsPath, _ := strings.CutSuffix(cfg.AssetsPath, "/")
    assetsPath, _ = strings.CutSuffix(assetsPath, "assets")
    http.Handle("/assets/", http.FileServer(http.Dir(assetsPath)))

    http.HandleFunc("/", makeGzipHandler(srv.rootHandler))
    http.HandleFunc("/autocomplete", srv.autocompleteHandler)
    http.HandleFunc("/search", makeGzipHandler(srv.searchHandler))
    http.HandleFunc("/articleClick", srv.articleClickHandler)

    log.Printf("sever started in %v", time.Since(start))
    err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
    srv.stop()
    log.Fatal(err)
}

func (s *server) stop() {
    log.Println("stopping server...")
}

func (s *server) rootHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{"Query": "", "SearchResults":""}
    buf, err := s.renderFullPage(data)
    if err != nil {
        log.Printf("ERROR: / rendering full page %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    if _, err = buf.WriteTo(w); err != nil {
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
    searchResults, err := s.fts.Search(query, s.resultsPerPage)
    if err != nil {
        log.Printf("ERROR: /serach error %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Header().Set("hx-push-url", fmt.Sprintf("/search?query=%s", query))
    w.Header().Set("hx-history-restore", "true")

    var buf bytes.Buffer
    buf, err = s.renderPartialSearchResults(searchResults)
    if err != nil {
        log.Printf("ERROR: /search partial render %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return 
    }

    if !isPartialRequest(r.Header) {
        buf, err = s.renderFullSearchResults(query, buf, searchResults)
    }
    if err != nil {
        log.Printf("ERROR: /search full render %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return 
    }

    if _, err := buf.WriteTo(w); err != nil {
        log.Printf("ERROR: /search writing response %s", err)
        w.WriteHeader(http.StatusInternalServerError)
        return 
    }
}

func (s *server) renderFullSearchResults(query string, partialResults bytes.Buffer, searchResults []FullTextSearchResult) (bytes.Buffer, error) {
    data := map[string]string{"Query": query, "SearchResults": partialResults.String()}
    return s.renderFullPage(data)
}

func (s *server) renderPartialSearchResults(searchResults []FullTextSearchResult) (bytes.Buffer, error) {
    return s.render("./templates/search_results.html", searchResults)
}

func (s *server) renderFullPage(data map[string]string) (bytes.Buffer, error) {
    return s.render("./templates/index.html", data)
}

func (s *server) render(templatePath string, data any) (bytes.Buffer, error) {
    var res bytes.Buffer
    tmpl, err := template.ParseFiles(templatePath)
    if err != nil {
        return res, err
    }
    err = tmpl.Execute(&res, data);
    return res, err
}

func (s *server) articleClickHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    article := r.URL.Query().Get("article")
    if article == "" {
        log.Println("WARN: /articleClick got no article")
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    id, err := strconv.Atoi(article)
    if err != nil {
        log.Printf("ERROR: /articleClick %s", err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    url, err := s.fts.GetUrl(id)
    if err != nil {
        log.Printf("ERROR: /articleClick failed to get url %s", err)
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    w.Header().Add("HX-Redirect", url)
    w.WriteHeader(http.StatusOK)
}

func isPartialRequest(header http.Header) bool {
    return header.Get("Hx-Request") == "true"
}


// TODO move to utils
type gzipResponseWriter struct {
    http.ResponseWriter
    buf bytes.Buffer
}
 
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
    return w.buf.Write(b)
}
 
func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            fn(w, r)
            return
        }
        w.Header().Set("Content-Encoding", "gzip")
        gzw := &gzipResponseWriter{ResponseWriter: w}
        fn(gzw, r)
        w.Header().Set("Content-Type", http.DetectContentType(gzw.buf.Bytes()))
        gz := gzip.NewWriter(w)
        defer gz.Close()
        gzw.buf.WriteTo(gz)
    }
}
