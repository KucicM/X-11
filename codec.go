package main

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Codec struct {
    id int
    inferenceOnly bool
    decodeCache map[int]string
    encodeCache map[string]int
}

func NewCodec(inferenceOnly bool) *Codec {
    codec := &Codec{
        id: 1, 
        inferenceOnly: inferenceOnly, 
        decodeCache: make(map[int]string), 
        encodeCache: make(map[string]int),
    }

    if codec.inferenceOnly {
        codec.populateCache()
    } else {
        log.Println("rebuilding terms.db")
        db := sqlx.MustOpen("sqlite3", "terms.db")
        defer db.Close()
        db.MustExec("CREATE TABLE IF NOT EXISTS terms (id INTEGER, term VARCHAR(30));")
        db.MustExec("DELETE FROM terms")
        db.MustExec("PRAGMA synchronous = OFF;")
        db.MustExec("PRAGMA journal_mode = MEMORY;")
    }

    return codec
}

func (c *Codec) Decode(val int) string {
    return c.decodeCache[val]
}

func (c *Codec) Encode(val string) int {
    id, ok := c.encodeCache[val]
    if !ok && ! c.inferenceOnly {
        id, c.id = c.id, c.id + 1
    }
    return id
}

func (c *Codec) populateCache() {
    db := sqlx.MustOpen("sqlite3", "terms.db")
    defer db.Close()

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
        c.encodeCache[term] = id
        c.decodeCache[id] = term
    }
}

func (m *Codec) Save() {
    log.Println("saving terms mapping")
    defer func(start time.Time) {
        log.Printf("terms mapping saved in %v", time.Since(start))
    }(time.Now())

    db := sqlx.MustOpen("sqlite3", "terms.db")
    defer db.Close()

    tx := db.MustBegin()
    stmt, err := tx.Preparex("INSERT INTO terms (id, term) VALUES ($1, $2);")
    if err != nil {
        log.Fatalf("ERROR: failed to create prepare %s", err)
    }

    for term, id := range m.encodeCache {
        _ = stmt.MustExec(id, term)
    }

    if err := tx.Commit(); err != nil {
        log.Fatalf("ERROR: failed to commit to term db")
    }
}
