package main

import (
	"container/heap"
	"log"
	"math"
	"sort"
	"time"
)

type TfIdf struct {
    FileMap map[string]*FileData `json:"fileMap"`
    Idf map[string]float64 `json:"idf"`
}

func NewTfIdf() *TfIdf {
    return &TfIdf{
        FileMap: make(map[string]*FileData),
        Idf: make(map[string]float64),
    }
}

func (t *TfIdf) computerIdf(term string) float64 {
    totalDocumnetCount := float64(len(t.FileMap))
    absTermFreq := t.Idf[term]
    return max(1.0, math.Log10(totalDocumnetCount / (1 + absTermFreq)))
}

type FileData struct {
    Name string `json:"name"`
    TermFreq map[string]float64 `json:"termFreq"`
    TotalCount float64 `json:"TotalCount"`
}

func (fd *FileData) computeTf(term string) float64 {
    if count, ok := fd.TermFreq[term]; ok{
        return count / fd.TotalCount
    }
    return 0
}

func (fd *FileData) Inc(term string) {
    fd.TotalCount += 1
    if val, ok := fd.TermFreq[term]; ok {
        fd.TermFreq[term] = val + 1
    } else {
        fd.TermFreq[term] = 1
    }
}

func (t *TfIdf) Insert(name string, ngram Ngram) {
    if _, ok := t.FileMap[name]; !ok {
        t.FileMap[name] = &FileData{Name: name, TermFreq: make(map[string]float64)}
    }

    term := createTerm(ngram)

    val := t.Idf[term]
    t.Idf[term] = val + 1

    data := t.FileMap[name]
    data.Inc(term)
}


type tfIdfResult []SearchResult

func (r tfIdfResult) Len() int {
    return len(r)
}

func (r tfIdfResult) Less(i, j int) bool {
    return r[i].Rank < r[j].Rank
}

func (r tfIdfResult) Swap(i, j int) {
    r[i], r[j] = r[j], r[i]
}

func (r *tfIdfResult) Push(it any) {
    *r = append(*r, it.(SearchResult))
}

func (r *tfIdfResult) Pop() any {
    old := *r
    n := len(old)
    x := old[n-1]
    *r = old[:n-1]
    return x
}

func (r tfIdfResult) Top() float64 {
    return r[0].Rank
}


func (t *TfIdf) Search(tokens []Token, maxResult int) []SearchResult {
    start := time.Now()
    out := new(tfIdfResult)

    for fileName, v := range t.FileMap {
        var totalRank = 0.0
        for _, token := range tokens {
            term := string(token)
            totalRank += v.computeTf(term) * t.computerIdf(term)
        }

        if (out.Len() <= maxResult || out.Top() < totalRank) && totalRank > 0 {
            heap.Push(out, SearchResult{fileName, totalRank})
        }


        for out.Len() > maxResult {
            heap.Pop(out)
        }
    }

    // high to low sort
    sort.Slice(*out, func(i, j int) bool {
        return !out.Less(i, j)
    })

    log.Printf("found in %v", time.Since(start))
    return *out
}

func createTerm(ngram Ngram) string {
    out := make([]rune, 0)
    out = append(out, ngram[0]...)
    for i := 1; i < len(ngram); i++ {
        out = append(out, ' ')
        out = append(out, ngram[i]...)
    }
    return string(out)
}

// todo handle errors
func (t *TfIdf) Save(path string) error {
    log.Printf("saving tf-idf to %s", path)
    start := time.Now()
    //file, _ := json.MarshalIndent(t.FileMap, " ", "")
   // _ = os.WriteFile("search_index.json", file, 0644)
    log.Printf("saving tf-idf took %v", time.Since(start))
    return nil
}

