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
}

func newFullTextIndex(cfg FullTextIndexCfg) *fullTextIndex {
	db := sqlx.MustOpen("sqlite3", cfg.DbFilePath)

	var initQueries = []string{
		"DROP TABLE IF EXISTS tf_idf_index;",
		"DROP TABLE IF EXISTS files;",
		"CREATE TABLE IF NOT EXISTS tf_idf_index (token_id INTEGER, tf REAL, idf REAL, file_id INTEGER);",
		"CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, file_name varchar(255), file_path varchar(255));",
		"PRAGMA synchronous = OFF;",
		"PRAGMA journal_mode = MEMORY;",
	}

	for _, query := range initQueries {
		if _, err := db.Exec(query); err != nil {
			db.Close()
			log.Fatalf("ERROR: exec on query '%s' failes %s", query, err)
		}
	}

	return &fullTextIndex{
		db:     db,
		idf:    &idf{count: 0, tokenIdFreq: make(map[int]int)},
	}
}

func (b *fullTextIndex) AddDocument(file_name, file_path string, tokens []common.Token) {
	relativeFreq := computeRelativeFrequency(tokens)

	b.idf.update(relativeFreq)
	b.save(file_name, file_path, relativeFreq, len(tokens))
}

func computeRelativeFrequency(tokens []common.Token) map[int]int {
	relativeFreq := make(map[int]int)
	for _, token := range tokens {
		relativeFreq[token.Id] += 1
	}
	return relativeFreq
}

func (b *fullTextIndex) save(file_name, file_path string, relativeFreq map[int]int, tokenCount int) {
	tx := b.db.MustBegin()
	res := tx.MustExec("INSERT INTO files (file_name, file_path) VALUES ($1, $2)", file_name, file_path)
	fileId, err := res.LastInsertId()
	if err != nil {
		log.Fatalf("ERROR: failed to get last id %s", err)
	}

	stmt, err := tx.Preparex("INSERT INTO tf_idf_index (token_id, tf, file_id) VALUES ($1, $2, $3);")
	if err != nil {
		log.Fatalf("ERROR: failed to prepare stmt %s", err)
	}
	defer stmt.Close()

	for tokenId, freq := range relativeFreq {
		tf := float64(freq) / float64(tokenCount)
		_ = stmt.MustExec(tokenId, tf, fileId)
	}

	if err = tx.Commit(); err != nil {
		log.Fatalf("ERROR: cannot commit %s", err)
	}
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

	stmt, err := tx.Preparex("UPDATE tf_idf_index SET idf = $1 WHERE token_id = $2")
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
	tokenIdFreq map[int]int
}

func (i *idf) update(relativeFreq map[int]int) {
	for k, v := range relativeFreq {
		i.tokenIdFreq[k] += v
		i.count += v
	}
}

func (i *idf) get() map[int]float64 {
	out := make(map[int]float64)
	for tokenId, freq := range i.tokenIdFreq {
		out[tokenId] = max(1.0, math.Log10(float64(i.count)/float64(1+freq)))
	}
	return out
}
