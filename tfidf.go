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

// add index
type SearchIndexBuilder struct {
    db *sql.DB

    // used for IDF
    totalTermCount int
    absTermFreq map[string]int

    // cache mappings
    nextTermId int
    termToId map[string]int
}

func NewSearchIndexBuilder() (*SearchIndexBuilder, error) {
    db, err := sql.Open("sqlite3", "test.db")
    if err != nil {
        return nil, err
    }

    log.Println("createing tf-idf table")
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS tf_idf_index (
        term_id INTEGER
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

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS terms (
        id INTEGER
        , term VARCHAR(30)
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

    log.Println("truncating file data")
    _, err = db.Exec("DELETE FROM terms;")
    if err != nil {
        return nil, err
    }

    return &SearchIndexBuilder{
        db: db, 
        totalTermCount: 0, 
        absTermFreq: make(map[string]int),
        nextTermId: 1,
        termToId: make(map[string]int),
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

    stmt, err := tx.Prepare("INSERT INTO tf_idf_index (term_id, tf, file_id) VALUES ($1, $2, $3);")
    if err != nil {
        log.Println(err)
        return 
    }
    defer stmt.Close()

    for term, freq := range relativeFreq {
        tf := float64(freq) / float64(totalFreq)
        termId := b.getTermId(term)
        _, err := stmt.Exec(termId, tf, fileId)
        if err != nil {
            log.Println(err)
        }
    }

    if err = tx.Commit(); err != nil {
        log.Println(err)
    }
}

func (b *SearchIndexBuilder) getTermId(term string) int {
    if _, ok := b.termToId[term]; !ok {
        b.termToId[term] = b.nextTermId
        b.nextTermId++
    }
    return b.termToId[term]
}

func (b *SearchIndexBuilder) Close() {
    log.Println("saving idf")
    if err := b.saveIDF(); err != nil {
        log.Println(err)
    }

    if err := b.saveTermIds(); err != nil {
        log.Println(err)
    }
}

func (b *SearchIndexBuilder) saveTermIds() error {
    tx, err := b.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("INSERT INTO terms (id, term_id) VALUES ($1, $2);")
    if err != nil {
        return err
    }

    for term, id := range b.termToId {
        if _, err := stmt.Exec(id, term); err != nil {
            return err
        }
    }

    return tx.Commit()

}

func (b *SearchIndexBuilder) saveIDF() error {
    tx, err := b.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("UPDATE tf_idf_index SET idf = $1 WHERE term = $2")
    if err != nil {
        return err
    }

    for term, freq := range b.absTermFreq {
        idf := max(1.0, math.Log10(float64(b.totalTermCount) / float64(1 + freq)))
        if _, err := stmt.Exec(term, idf); err != nil {
            return err
        }
    }

    return tx.Commit()
}

