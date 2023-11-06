package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// inference
type searchTerm string

type SearchIndex struct {
    db *sql.DB // make connection pool

    termToId map[string]int
}

func NewSeachIndex() (*SearchIndex, error){
    db, err := sql.Open("sqlite3", "test.db")
    if err != nil {
        return nil, err
    }

    // load termIds
    _, err = db.Query("SELECT id, term FROM terms;")
    if err != nil {
        return nil, err
    }

    return &SearchIndex{db: db}, nil
}

// maybe seperate TF and IDF in two queries per term and cache most common used terms?
func (i *SearchIndex) Search(tokens []Token, maxReturn int) []SearchResult {
    out := make([]SearchResult, 0)

    term_ids := make([]int, 0, len(tokens))
    for _, term := range toSeachTerms(tokens) {
        if id, ok := i.termToId[term]; ok {
            term_ids = append(term_ids, id)
        }
    }

    predicate := strings.Repeat("?,", len(term_ids)-1) + "?"
    query := fmt.Sprintf(`
    SELECT file_id, SUM(tf * idf) tf_idf
    FROM ft_idf_index
    WHERE term_id in (%s)
    GROUP BY file_id
    HAVING tf_idf > 0
    ORDER BY tf_idf DESC;
    `, predicate)
    rows, err := i.db.Query(query, term_ids)
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

func toSeachTerms(tokens []Token) []string {
    terms := make([]string, 0, len(tokens))
    for _, token := range tokens {
        terms = append(terms, string(token))
    }
    return terms
}


// creating index
var initQueries = []string{
    "CREATE TABLE IF NOT EXISTS tf_idf_index (term_id INTEGER, tf REAL, idf REAL, file_id INTEGER);",
    "CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, file_name varchar(255), file_path varchar(255));",
    "CREATE TABLE IF NOT EXISTS terms (id INTEGER, term VARCHAR(30));",
    "DELETE FROM tf_idf_index;",
    "DELETE FROM files;",
    "DELETE FROM terms;",
    "DROP INDEX IF EXISTS idx_term_id_tf_idf_index;",
}
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

    for _, query := range initQueries {
        if _, err := db.Exec(query); err != nil {
            db.Close()
            return nil, fmt.Errorf("ERROR: running query `%s` resulted in error %s", query, err)
        }
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
    defer b.db.Close()
    log.Println("saving cached values")

    if err := b.createIndices(); err != nil {
        log.Println(err)
    }

    if err := b.saveIDF(); err != nil {
        log.Println(err)
    }

    if err := b.saveTermIds(); err != nil {
        log.Println(err)
    }
}

func (b *SearchIndexBuilder) createIndices() error {
    log.Println("indexing db...")
    defer func(start time.Time) {
        log.Printf("db indexing done in %v", time.Since(start))
    }(time.Now())

    if _, err := b.db.Exec("CREATE INDEX idx_term_id_tf_idf_index ON tf_idf_index (term_id);"); err != nil {
        return err
    }

    return nil
}

// TODO should probably be part of tokenizer
func (b *SearchIndexBuilder) saveTermIds() error {
    log.Println("saving terms mapping")
    defer func(start time.Time) {
        log.Printf("terms mapping saved in %v", time.Since(start))
    }(time.Now())

    tx, err := b.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("INSERT INTO terms (id, term) VALUES ($1, $2);")
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
    log.Println("saving idf")
    defer func(start time.Time) {
        log.Printf("idf saved in %v", time.Since(start))
    }(time.Now())

    tx, err := b.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("UPDATE tf_idf_index SET idf = $1 WHERE term_id = $2")
    if err != nil {
        return err
    }

    for term, freq := range b.absTermFreq {
        idf := max(1.0, math.Log10(float64(b.totalTermCount) / float64(1 + freq)))
        termId := b.getTermId(term)
        if _, err := stmt.Exec(termId, idf); err != nil {
            return err
        }
    }

    return tx.Commit()
}

