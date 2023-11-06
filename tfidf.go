package main

import (
	"container/heap"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

    _ "github.com/mattn/go-sqlite3"
)

// inference
type searchTerm string

type SearchIndex struct {
    db *sql.DB // make connection pool
}

func NewSeachIndex() (*SearchIndex, error){
    db, err := sql.Open("sqlite3", "test.db")
    if err != nil {
        return nil, err
    }

    return &SearchIndex{db: db}, nil
}

// maybe seperate TF and IDF in two queries per term and cache most common used terms?
func (i *SearchIndex) Search(tokens []Token, maxReturn int) []SearchResult {
    out := make([]SearchResult, 0)

    terms := toSeachTerms(tokens)

    predicate := strings.Repeat("?,", len(terms)-1) + "?"
    query := fmt.Sprintf(`
        SELECT 
            f.FileName
            , SUM(f.TF * g.IDF) TfIdf
        FROM FileIndex f
        JOIN GlobalIndex g ON g.Term = f.Term
        WHERE f.term in (%s)
        GROUP BY f.FileName
        HAVING TfIdf > 0
        ORDER BY TfIdf;
    `, predicate)
    rows, err := i.db.Query(query, terms)
    if err != nil {
        log.Println(err)
    }

    for rows.Next() {
        var row SearchResult
        if err := rows.Scan(&row.Title, &row.Rank); err != nil {
            log.Println(err)
        }
        out = append(out, row)
    }

    return out
}

func toSeachTerms(tokens []Token) []searchTerm {
    terms := make([]searchTerm, 0, len(tokens))
    for _, token := range tokens {
        terms = append(terms, searchTerm(token))
    }
    return terms
}

// build index
// take whole document in memory
// build TF
// save TF to DB
// update in memory IDF
// after all documents save IDF to DB

// add index

type SearchIndexBuilder struct {
    db *sql.DB

    // used for IDF
    totalTermCount int
    absTermFreq map[string]int
}

func NewSearchIndexBuilder() (*SearchIndexBuilder, error) {
    db, err := sql.Open("sqlite3", "test.db")
    if err != nil {
        return nil, err
    }

    // todo maybe move name to other table to reduce size
    log.Println("createing tf-idf table")
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS tf_idf_index (
        term VARCHAR(30)
        , tf REAL
        , idf REAL
        , file_id INTEGER
    );`)
    if err != nil {
        return nil, err
    }

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY
        , file_name varchar(255)
        , file_path varchar(255)
    );`)
    if err != nil {
        return nil, err
    }

    log.Println("truncating tf-idf index")
    _, err = db.Exec("DELETE FROM tf_idf_index;")
    if err != nil {
        return nil, err
    }

    log.Println("truncating file data")
    _, err = db.Exec("DELETE FROM files;")
    if err != nil {
        return nil, err
    }

    return &SearchIndexBuilder{
        db: db, 
        totalTermCount: 0, 
        absTermFreq: make(map[string]int),
    }, nil
}

func (b *SearchIndexBuilder) AddDocument(file_name, file_path string, document []Token) {
    relativeFreq := make(map[string]int)
    for _, token := range document {
        relativeFreq[string(token)] += 1
    }

    // idf part
    totalFreq := 0
    for k, v := range relativeFreq {
        totalFreq += v
        b.absTermFreq[k] += v
    }
    b.totalTermCount += totalFreq

    // tf part
    tx, err := b.db.Begin()
    if err != nil {
        log.Println(err)
        return 
    }

    res, err := tx.Exec("INSERT INTO files (file_name, file_path) VALUES ($1, $2)", file_name, file_path)
    if err != nil {
        log.Println(err)
        return
    }
    fileId, err := res.LastInsertId()
    if err != nil {
        log.Println(err)
        return
    }

    stmt, err := tx.Prepare("INSERT INTO tf_idf_index (term, tf, file_id) VALUES ($1, $2, $3);")
    if err != nil {
        log.Println(err)
        return 
    }
    defer stmt.Close()

    for term, freq := range relativeFreq {
        tf := float64(freq) / float64(totalFreq)
        _, err := stmt.Exec(term, tf, fileId)
        if err != nil {
            log.Println(err)
        }
    }

    if err = tx.Commit(); err != nil {
        log.Println(err)
    }
}

func (b *SearchIndexBuilder) Close() {
    log.Println("saving idf")
    tx, err := b.db.Begin()
    if err != nil {
        log.Println(err)
        return
    }

    stmt, err := tx.Prepare("UPDATE tf_idf_index SET idf = $1 WHERE term = $2")
    if err != nil {
        log.Println(err)
        return
    }

    for term, freq := range b.absTermFreq {
        idf := max(1.0, math.Log10(float64(b.totalTermCount) / float64(1 + freq)))
        _, err := stmt.Exec(term, idf)
        if err != nil {
            log.Println(err)
        }
    }
    if err = tx.Commit(); err != nil {
        log.Println(err)
    }
}

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

    // for token in tokens
    // calculate idf
    // find files with token
    // calculate tf of all files
    // calculate tf-idf
    // sort
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

