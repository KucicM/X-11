package main

import (
	"log"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// inference
type SearchIndex struct {
    mapper *termMapper
}

func NewSeachIndex() *SearchIndex{
    mapper := LoadMapper()
    return &SearchIndex{mapper: mapper}
}

// maybe precompute whole table?
func (i *SearchIndex) Search(tokens []string, maxReturn int) ([]SearchResult, error) {
    defer func(start time.Time) {
        log.Printf("INFO: %d terms in %v", len(tokens), time.Since(start))
    }(time.Now())

    termIds := i.mapper.precomputedIds(tokens)
    if len(termIds) == 0 {
        return []SearchResult{}, nil
    }

   db, err := sqlx.Open("sqlite3", "test.db?mode=ro")
   if err != nil {
       return nil, err
   }
   defer db.Close()

    q := `SELECT f.file_name, SUM(i.tf * i.idf) rank
    FROM tf_idf_index i
    JOIN files f ON f.id = i.file_id
    WHERE i.term_id IN (?)
    GROUP BY i.file_id
    HAVING rank > 0
    ORDER BY rank DESC
    LIMIT 20;`
    query, args, err := sqlx.In(q, termIds)
    if err != nil {
        return nil, err
    }

    var out []SearchResult
    if err := db.Select(&out, query, args...); err != err {
        return nil, err
    }

    return out, nil
}

// creating index
type idf struct {
    count int
    termIdFreq map[int]int
}

func (i *idf) update(relativeFreq map[int]int) {
    for k, v := range relativeFreq {
        i.termIdFreq[k] += v
        i.count += v
    }
}

func (i *idf) get() map[int]float64 {
    out := make(map[int]float64)
    for termId, freq := range i.termIdFreq {
        out[termId] = max(1.0, math.Log10(float64(i.count) / float64(1 + freq)))
    }
    return out
}

type SearchIndexBuilder struct {
    db *sqlx.DB
    idf *idf

    // cache mappings TODO remove
    mapper *termMapper
}

func NewSearchIndexBuilder() *SearchIndexBuilder {
    db := sqlx.MustOpen("sqlite3", "test.db")

    var initQueries = []string{
        "CREATE TABLE IF NOT EXISTS tf_idf_index (term_id INTEGER, tf REAL, idf REAL, file_id INTEGER);",
        "CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, file_name varchar(255), file_path varchar(255));",
        "DELETE FROM tf_idf_index;",
        "DELETE FROM files;",
        "DROP INDEX IF EXISTS idx_term_id_tf_idf_index;",
        "PRAGMA synchronous = OFF;",
        "PRAGMA journal_mode = MEMORY;",
    }

    for _, query := range initQueries {
        if _, err := db.Exec(query); err != nil {
            db.Close()
            log.Fatalf("ERROR: exec on query '%s' failes %s", query, err)
        }
    }

    return &SearchIndexBuilder{
        db: db, 
        idf: &idf{count: 0, termIdFreq: make(map[int]int)},
        mapper: BuildMapper(),
    }
}

func (b *SearchIndexBuilder) AddDocument(file_name, file_path string, document []string) {
    termIds := b.mapper.computeIds(document)
    relativeFreq := computeRelativeFrequency(termIds)

    b.idf.update(relativeFreq)
    b.save(file_name, file_path, relativeFreq, len(document))
}

func computeRelativeFrequency(termIds []int) map[int]int {
    relativeFreq := make(map[int]int)
    for _, termId := range termIds {
        relativeFreq[termId] += 1
    }
    return relativeFreq
}

func (b *SearchIndexBuilder) save(file_name, file_path string, termIdFreq map[int]int, tokenCount int) {
    tx := b.db.MustBegin()
    res := tx.MustExec("INSERT INTO files (file_name, file_path) VALUES ($1, $2)", file_name, file_path)
    fileId, err := res.LastInsertId()
    if err != nil {
        log.Fatalf("ERROR: failed to get last id %s", err)
    }


    stmt, err := tx.Preparex("INSERT INTO tf_idf_index (term_id, tf, file_id) VALUES ($1, $2, $3);")
    if err != nil {
        log.Fatalf("ERROR: failed to prepare stmt %s", err)
    }
    defer stmt.Close()

    for termId, freq := range termIdFreq {
        tf := float64(freq) / float64(tokenCount)
        _ = stmt.MustExec(termId, tf, fileId)
    }

    if err = tx.Commit(); err != nil {
        log.Fatalf("ERROR: cannot commit %s", err)
    }
}

func (b *SearchIndexBuilder) Close() {
    defer b.db.Close()
    log.Println("saving cached values")
    b.createIndices()
    b.saveIDF()
    b.mapper.Save()
}

func (b *SearchIndexBuilder) createIndices() {
    log.Println("indexing db...")
    defer func(start time.Time) {
        log.Printf("db indexing done in %v", time.Since(start))
    }(time.Now())

    b.db.MustExec("CREATE INDEX idx_term_id_tf_idf_index ON tf_idf_index (term_id);")
}

func (b *SearchIndexBuilder) saveIDF() {
    log.Println("saving idf")
    defer func(start time.Time) {
        log.Printf("idf saved in %v", time.Since(start))
    }(time.Now())

    tx := b.db.MustBegin()

    stmt, err := tx.Preparex("UPDATE tf_idf_index SET idf = $1 WHERE term_id = $2")
    if err != nil {
        log.Fatalf("ERROR: failed to create prepared statment for idf save %s", err)
    }
    defer stmt.Close()

    for termId, idf := range b.idf.get() {
        _ = stmt.MustExec(idf, termId)
    }

    if err := tx.Commit(); err != nil {
        log.Fatalf("ERROR: failed to commit idf save %s", err)
    }
}

// TODO: move this part to lexer
type termMapper struct {
    id int
    termToId map[string]int
    idToTerm map[int]string
}

func LoadMapper() *termMapper {
    db := sqlx.MustOpen("sqlite3", "terms.db")
    defer db.Close()

    termToId := make(map[string]int)
    idToTerm := make(map[int]string)
    rows, err := db.Query("SELECT id, term FROM terms;")
    if err != nil {
        log.Fatalf("ERROR: cannot query terms %s", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var term string
        if err := rows.Scan(&id, &term); err != nil {
            log.Fatalf("ERROR: cannot scan term row %s", err)
        }
        termToId[term] = id
        idToTerm[id] = term
    }

    return &termMapper{termToId: termToId, idToTerm: idToTerm}
}

func BuildMapper() *termMapper {
    db := sqlx.MustOpen("sqlite3", "terms.db")
    defer db.Close()

    // ddl
    db.MustExec("CREATE TABLE IF NOT EXISTS terms (id INTEGER, term VARCHAR(30));")
    db.MustExec("DELETE FROM terms")
    db.MustExec("PRAGMA synchronous = OFF;")
    db.MustExec("PRAGMA journal_mode = MEMORY;")

    return &termMapper{id: 1, termToId: make(map[string]int), idToTerm: make(map[int]string)}
}

func (m * termMapper) computeId(term string) int {
    if _, ok := m.termToId[term]; !ok {
        m.id += 1
        m.termToId[term] = m.id
    }
    return m.termToId[term]
}

func (m *termMapper) computeIds(terms []string) []int {
    out := make([]int, 0, len(terms))
    for _, term := range terms {
        out = append(out, m.computeId(term))
    }
    return out
}

func (m *termMapper) precomputedIds(terms []string) []int {
    out := make([]int, 0, len(terms))
    for _, term := range terms {
        if id, ok := m.termToId[term]; ok {
            out = append(out, id)
        }
    }
    return out
}

func (m *termMapper) Save() {
    log.Println("saving terms mapping")
    defer func(start time.Time) {
        log.Printf("terms mapping saved in %v", time.Since(start))
    }(time.Now())

    db := sqlx.MustOpen("sqlite3", "terms.db")
    defer db.Close()

    tx := db.MustBegin()

    tx.MustExec("CREATE TABLE IF NOT EXISTS terms (id INTEGER, term VARCHAR(30));")
    tx.MustExec("DELETE FROM terms")
    tx.MustExec("PRAGMA synchronous = OFF;")
    tx.MustExec("PRAGMA journal_mode = MEMORY;")

    stmt, err := tx.Preparex("INSERT INTO terms (id, term) VALUES ($1, $2);")
    if err != nil {
        log.Fatalf("ERROR: failed to create prepare %s", err)
    }

    for term, id := range m.termToId {
        _ = stmt.MustExec(id, term)
    }

    if err := tx.Commit(); err != nil {
        log.Fatalf("ERROR: failed to commit to term db")
    }
}
