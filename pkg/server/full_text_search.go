package server

import (
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kucicm/X-11/pkg/common"
)

type Document struct {
    Title string
    Tokens []string
    Url string
}

type FullTextSearchCfg struct {
    DbUrl string `json:"db-url"`
    Rebuild bool `json:"rebuild"`
    ResultsPerPage int `json:"results-per-page"`
}

type FullTextSearchResult struct {
    FileId int `db:"file_id"`
    Title string `db:"title"`
    Description string `db:"description"`
    NextPage string // should not be here...
}

type FullTextSearch struct {
    dbUrl string
    limit int
}

func NewFullTextSearch(cfg FullTextSearchCfg) *FullTextSearch {
    fts := &FullTextSearch{dbUrl: cfg.DbUrl, limit: cfg.ResultsPerPage}
    if cfg.Rebuild {
        fts.rebuild()
    }
    return fts
}

func (fts *FullTextSearch) rebuild() {
    db := sqlx.MustOpen("sqlite3", fts.dbUrl)

    log.Println("rebuilding full text search")
    _ = db.MustExec("DROP TABLE IF EXISTS documents;")
    _ = db.MustExec("DROP TABLE IF EXISTS texts;")
    _ = db.MustExec(`CREATE TABLE documents (
        id INTEGER PRIMARY KEY
        , url TEXT
        , title TEXT
    );`)
    _ = db.MustExec(`CREATE VIRTUAL TABLE texts USING FTS5(
        , text
        , tokenize="PORTER ASCII"
    );`)
    if err := db.Close(); err != nil {
        log.Fatalf("ERROR: cannot close fts db %s", err)
    }
}

func (fts *FullTextSearch) AddDocument(doc Document) {
    db := sqlx.MustOpen("sqlite3", fts.dbUrl)
    defer db.Close()

    _ = db.MustExec("INSERT INTO documents (url, title) VALUES ($1, $2);", doc.Url, doc.Title)

    text := strings.Join(doc.Tokens, " ")
    _ = db.MustExec("INSERT INTO texts (text) VALUES (?);", text)
}

func (fts *FullTextSearch) FinishIndexing() {
    db := sqlx.MustOpen("sqlite3", fts.dbUrl)
    defer db.Close()

    _ = db.MustExec("VACUUM;")
}

func (fts *FullTextSearch) Search(query string, page int) ([]FullTextSearchResult, error) {
    tokens := common.Tokenize(query)
    if len(tokens) == 0 {
        return make([]FullTextSearchResult, 0), nil
    }

    db := sqlx.MustOpen("sqlite3", fts.dbUrl)
    defer func(start time.Time) {
        db.Close()
        log.Printf("INFO: search of %d tokens took %s", len(tokens), time.Since(start))
    }(time.Now())

    selectQuery := `
    SELECT 
        f.id as file_id
        , title 
        , SNIPPET(texts, 0, '', '', '...', 40) as description
    FROM texts as t
    JOIN documents as f on f.id = t.rowid
    WHERE text MATCH $1
    ORDER BY rank
    LIMIT $2
    OFFSET $3;`

    query = strings.Join(tokens, " ")
    offset := fts.computeOffset(page)
    var res []FullTextSearchResult
    if err := db.Select(&res, selectQuery, query, fts.limit, offset); err != nil {
        return nil, err
    }

    return res, nil
}

func (fts *FullTextSearch) GetUrl(file_id int) (string, error) {
    db := sqlx.MustOpen("sqlite3", fts.dbUrl)
    defer func(start time.Time) {
        db.Close()
        log.Printf("INFO: fetch url in %s", time.Since(start))
    }(time.Now())

    var url string
    if err := db.QueryRow("SELECT url FROM documents WHERE id = $1", file_id).Scan(&url); err != nil {
        return "", err
    }
    return url, nil
}

func (fts *FullTextSearch) computeOffset(page int) int {
    return page * fts.limit
}

