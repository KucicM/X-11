// build full text search index and save to database
package build

import (
	"log"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kucicm/X-11/pkg/common"
)

type FullTextIndexCfg struct {
    DbFilePath string `json:"db-file-path"`
}

type fullTextIndex struct {
	db  *sqlx.DB
	idf *idf
    tokenCache map[string]int64
}

func newFullTextIndex(cfg FullTextIndexCfg) *fullTextIndex {
	db := sqlx.MustOpen("sqlite3", cfg.DbFilePath)

	var initQueries = []string{
		"DROP TABLE IF EXISTS tf_idf_index;",
		"DROP TABLE IF EXISTS files;",
		"CREATE TABLE IF NOT EXISTS tf_idf_index (token_id INTEGER, tfidf REAL, file_id INTEGER);",
		"CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, title varchar(255), url varchar(255), description TEXT);",
		"PRAGMA synchronous = OFF;",
		"PRAGMA journal_mode = MEMORY;",
	}

	for _, query := range initQueries {
        db.MustExec(query)
	}

	return &fullTextIndex{
		db:     db,
		idf:    &idf{count: 0, tokenIdFreq: make(map[uint32]int)},
        tokenCache: make(map[string]int64),
	}
}

func (b *fullTextIndex) AddDocument(doc Document) {
	relativeFreq := computeRelativeFrequency(doc.Tokens)

	b.idf.update(relativeFreq)
	b.save(doc, relativeFreq)
}

func computeRelativeFrequency(tokens []string) map[uint32]int {
	relativeFreq := make(map[uint32]int)
	for i := range tokens {
		relativeFreq[common.HashToken(tokens[i])] += 1
	}
	return relativeFreq
}

func (b *fullTextIndex) save(doc Document, relativeFreq map[uint32]int) {
	tx := b.db.MustBegin()

	res := tx.MustExec("INSERT INTO files (title, url, description) VALUES ($1, $2, $3)", doc.Title, doc.Url, doc.Description)
	fileId, err := res.LastInsertId()
	if err != nil {
		log.Fatalf("ERROR: failed to get last id %s", err)
	}

	stmt, err := tx.Preparex("INSERT INTO tf_idf_index (token_id, tfidf, file_id) VALUES ($1, $2, $3);")
	if err != nil {
		log.Fatalf("ERROR: failed to prepare stmt %s", err)
	}
	defer stmt.Close()

	for tokenId, freq := range relativeFreq {
		tf := float64(freq) / float64(len(doc.Tokens))
		_ = stmt.MustExec(tokenId, tf, fileId)
	}

	if err = tx.Commit(); err != nil {
		log.Fatalf("ERROR: cannot commit %s", err)
	}
}

func (b *fullTextIndex) getTokenId(stmt *sqlx.Stmt, token string) int64 {
    if tokenId, ok := b.tokenCache[token]; ok {
        return tokenId
    }

    r := stmt.MustExec(token)
    tokenId, err := r.LastInsertId()
    if err != nil {
        log.Fatalf("ERROR: cannot get last insert row id for token %s", err)
    }
    b.tokenCache[token] = tokenId
    return tokenId
}

func (b *fullTextIndex) close() {
	defer b.db.Close()
	log.Println("saving cached values")
	b.createIndices()
	b.saveIDF()
	b.vacuum()
}

func (b *fullTextIndex) createIndices() {
	log.Println("indexing full text index db...")
	defer func(start time.Time) {
		log.Printf("indexing full text index db took %s", time.Since(start))
	}(time.Now())

	b.db.MustExec("CREATE INDEX idx_token_id_tf_idf_index ON tf_idf_index (token_id);")
}

func (b *fullTextIndex) saveIDF() {
	log.Println("saving idf")
	defer func(start time.Time) {
		log.Printf("idf saved in %v", time.Since(start))
	}(time.Now())

	tx := b.db.MustBegin()

	stmt, err := tx.Preparex("UPDATE tf_idf_index SET tfidf = tfidf * $1 WHERE token_id = $2")
	if err != nil {
		log.Fatalf("ERROR: failed to create prepared statment for idf save %s", err)
	}
	defer stmt.Close()

	for tokenId, idf := range b.idf.get() {
		_ = stmt.MustExec(idf, tokenId)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("ERROR: failed to commit idf save %s", err)
	}
}

func (b *fullTextIndex) vacuum() {
	log.Println("running vacumm on full text index")
	defer func(start time.Time) {
        log.Printf("vacumm on full text index took: %v", time.Since(start))
	}(time.Now())
	b.db.MustExec("VACUUM;")
}

type idf struct {
	count      int
	tokenIdFreq map[uint32]int
}

func (i *idf) update(relativeFreq map[uint32]int) {
	for k, v := range relativeFreq {
		i.tokenIdFreq[k] += v
		i.count += v
	}
}

func (i *idf) get() map[uint32]float64 {
	out := make(map[uint32]float64)
	for token, freq := range i.tokenIdFreq {
		out[token] = max(1.0, math.Log10(float64(i.count)/float64(1+freq)))
	}
	return out
}
