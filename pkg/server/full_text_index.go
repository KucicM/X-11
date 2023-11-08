package server

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kucicm/X-11/pkg/common"
)

type FullTextIndex interface {
    Search(tokens []common.Token, maxResults int) ([]SearchIndexResult, error)
}

type FullTextIndexCfg struct {
    DbFilePath string `json:"db-file-path"`
}

type tfIdf struct {
    dbFilePath string
}

func NewFullTextIndex(cfg FullTextIndexCfg) FullTextIndex {
    db := sqlx.MustOpen("sqlite3", fmt.Sprintf("%s?mode.ro", cfg.DbFilePath))
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("cannot ping full text index db %s", err)
    }
    return &tfIdf{cfg.DbFilePath}
}

func (t *tfIdf) Search(tokens []common.Token, maxResults int) ([]SearchIndexResult, error) {
   defer func(start time.Time) {
        log.Printf("INFO: %d terms in %v", len(tokens), time.Since(start))
    }(time.Now())

    tokenIds := make([]int, 0, len(tokens))
    for i := range tokens {
        tokenIds = append(tokenIds, tokens[i].Id)
    }

    db, err := sqlx.Open("sqlite3", fmt.Sprintf("%s?mode.ro", t.dbFilePath))
    if err != nil {
        return nil, err
    }
    defer db.Close()

    q := `SELECT f.file_name, SUM(i.tf * i.idf) rank
    FROM tf_idf_index i
    JOIN files f ON f.id = i.file_id
    WHERE i.token_id IN (?)
    GROUP BY i.file_id
    HAVING rank > 0
    ORDER BY rank DESC
    LIMIT ?;`

    query, args, err := sqlx.In(q, tokenIds, maxResults)
    if err != nil {
        return nil, err
    }

    log.Println(query)

    var out []SearchIndexResult
    if err := db.Select(&out, query, args...); err != err {
        return nil, err
    }
    return out, nil
}
