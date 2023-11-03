package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type SuggestionResult struct {
    Suggestion string
    // maybe add from history vs no history
}

type SearchResult struct {
    Title string
}

var trie *Trie

func main() {
    trie = NewTrie()
    trie.build([]string{
        "real test1",
        "real test2",
        "real test3",
    })
    
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
    w.Header().Add("Cache-Control", "private, max-age=3600")
    query := strings.TrimSpace(r.URL.Query().Get("query"))
    if query == "" {
        return
    }


    tmpl := template.Must(template.ParseFiles("./templates/suggestion_results.html"))
    suggestions := make([]SuggestionResult, 0)
    for _, s := range trie.finaAll(query) {
        suggestions = append(suggestions, SuggestionResult{s})
    }
    data := map[string][]SuggestionResult{
        "Suggestions": suggestions,
    }
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


// trie for every prefix cache most common results
type Trie struct {
    root *TrieNode
}

func NewTrie() *Trie {
    return &Trie{newTrieNode()}
}

func (t *Trie) build(words []string) {
    log.Println("building trie")
    defer func(started time.Time) {
        log.Printf("trie build in %v\n", time.Since(started))
    }(time.Now())
    var node *TrieNode
    for _, word := range words {
        node = t.root
        for _, b := range []byte(word) {
            node = node.getNext(b)
        }
        node.mark()
    }
}

func (t *Trie) finaAll(prefix string) []string {
    node := t.root
    // todo search should not create new nodes
    for _, b := range []byte(prefix) {
        node = node.getNext(b)
    }

    bRets := node.findAll()
    ret := make([]string, 0, len(bRets))

    bprefix := []byte(prefix)
    for _, bs := range bRets {
        buff := make([]byte, 0, len(bprefix) + len(bs))
        buff = append(buff, bprefix...)

        for i := len(bs)-1; i >= 0; i-- {
            buff = append(buff, bs[i])
        }
        ret = append(ret, string(buff))
    }
    return ret
}

type TrieNode struct {
    next map[byte]*TrieNode
    cnt uint
}

func newTrieNode() *TrieNode {
    return &TrieNode{make(map[byte]*TrieNode), 0}
}

func (t *TrieNode) getNext(b byte) *TrieNode {
    if val, ok := t.next[b]; ok {
        return val
    }
    val := newTrieNode()
    t.next[b] = val
    return val
}

func (t *TrieNode) mark() {
    t.cnt += 1
}

func (t *TrieNode) findAll() [][]byte{
    ret := make([][]byte, 0)
    if t.cnt > 0 {
        ret = append(ret, []byte(""))
    }

    for k, v := range t.next {
        for _, val := range v.findAll() {
            val = append(val, k)
            ret = append(ret, val)
        }
    }
    return ret
}
