package common

import (
	"fmt"
	"log"
	"time"
	"unicode"

	"github.com/jmoiron/sqlx"
)

type TokenizerCfg struct {
    DbPath string `json:"db-path"`
    BuildMode bool `json:"build"`
}

type Tokenizer struct {
    id int
    buildMode bool
    decodeCache map[int]string
    encodeCache map[string]int
    dbPath string
}

func NewTokenizer(cfg TokenizerCfg) *Tokenizer {
    tok := &Tokenizer{
        id: 1, 
        buildMode: cfg.BuildMode, 
        decodeCache: make(map[int]string), 
        encodeCache: make(map[string]int),
        dbPath: cfg.DbPath,
    }

    if tok.buildMode {
        log.Printf("rebuilding %s", cfg.DbPath)
        db := sqlx.MustOpen("sqlite3", cfg.DbPath)
        defer db.Close()
        db.MustExec("CREATE TABLE IF NOT EXISTS tokens (id INTEGER, token VARCHAR(30));")
        db.MustExec("DELETE FROM tokens")
        db.MustExec("PRAGMA synchronous = OFF;")
        db.MustExec("PRAGMA journal_mode = MEMORY;")
    } else {
        tok.populateCache()
    }

    return tok
}

// TODO; reduce number of conversions
func (t *Tokenizer) TokenizeStr(str string) []Token {
    return t.Tokenize([]byte(str))
}

// []byte -> []rune -> [][]rune -> token
func (t *Tokenizer) Tokenize(content []byte) []Token {
    terms := breakIntoTerms(content)
    return t.toTokens(terms)
}


func (t *Tokenizer) Gramify(content []byte, minN, maxN int) []Token {
    terms := breakIntoTerms(content)
    if len(terms) < minN {
        return make([]Token, 0)
    }

    ngmraTerms := make([][]rune, 0, len(terms))
    for l, h := 0, minN; h <= len(terms); {
        if (h - l) > maxN {
            l += 1
        }

        for ll := l; (h - ll) >= minN; ll++ {
            ngmraTerms = append(ngmraTerms, mergeTerms(terms[ll:h]))
        }

        if h <= len(terms) {
            h++
        }
    }
    return t.toTokens(ngmraTerms)
}

func (t *Tokenizer) toTokens(terms [][]rune) []Token {
    ret := make([]Token, 0, len(terms))
    for i := range terms {
        if id := t.encode(string(terms[i])); id != 0 {
            ret = append(ret, Token{Id: id, Runes: terms[i]})
        }
    }
    return ret
}

func (t *Tokenizer) encode(val string) int {
    id, ok := t.encodeCache[val]
    if !ok && t.buildMode {
        id, t.id = t.id, t.id + 1
        t.encodeCache[val] = id
        t.decodeCache[id] = val
    }
    return id
}

func (t *Tokenizer) decode(val int) string {
    return t.decodeCache[val]
}

func (t *Tokenizer) populateCache() {
    db := sqlx.MustOpen("sqlite3", fmt.Sprintf("%s?mode=ro", t.dbPath))
    defer db.Close()

    rows, err := db.Query("SELECT id, token FROM tokens;")
    if err != nil {
        log.Fatalf("ERROR: cannot query tokens %s", err)
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var token string
        if err := rows.Scan(&id, &token); err != nil {
            log.Fatalf("ERROR: cannot scan token row %s", err)
        }
        t.encodeCache[token] = id
        t.decodeCache[id] = token
    }

}

func (t *Tokenizer) Close() {
    log.Printf("saving token mapping with %d tokens", len(t.encodeCache))
    defer func(start time.Time) {
        log.Printf("token mapping saved in %v", time.Since(start))
    }(time.Now())

    db := sqlx.MustOpen("sqlite3", t.dbPath)
    defer db.Close()

    tx := db.MustBegin()
    stmt, err := tx.Preparex("INSERT INTO tokens (id, token) VALUES ($1, $2);")
    if err != nil {
        log.Fatalf("ERROR: failed to create prepare statment for token insert %s", err)
    }

    for token, id := range t.encodeCache {
        _ = stmt.MustExec(id, token)
    }

    if err := tx.Commit(); err != nil {
        log.Fatalf("ERROR: failed to commit to token db")
    }

}

func mergeTerms(terms [][]rune) []rune {
    total := 0
    for i := range terms {
        total += len(terms[i]) + 1
    }

    ret := make([]rune, total)
    for i := range terms {
        ret = append(ret, terms[i]...)
    }
    return ret
}

// document = convert []rune to [][]rune
func breakIntoTerms(content []byte) [][]rune {
    doc := []rune(string(content))
    terms := make([][]rune, 0)
    for len(doc) > 0 {
        doc = removeLeftPadding(doc, func(r rune) bool {
            return unicode.IsSpace(r) || unicode.IsPunct(r)
        })

        if len(doc) == 0 {
            break
        }

        var term []rune
        if unicode.IsDigit(doc[0]) {
            doc, term = takeUntil(doc, unicode.IsDigit)
        } else if unicode.IsLetter(doc[0]) {
            doc, term = takeUntil(doc, unicode.IsLetter)
        } else {
            doc, term = take(doc, 1)
        }
        terms = append(terms, term)
    }
    return terms
}

func removeLeftPadding(doc []rune, toRemove func(rune) bool) []rune {
    for len(doc) > 0 && toRemove(doc[0]) {
        doc = doc[1:]
    }
    return doc
}

func takeUntil(doc []rune, fn func(rune) bool) ([]rune, []rune) {
    n := 0
    for n < len(doc) && fn(doc[n]) {
        n++
    }
    return take(doc, n)
}

func take(doc []rune, n int) ([]rune, []rune) {
    for i := 0; i < n; i++ {
        doc[i] = unicode.ToLower(doc[i])
    }
    ret := doc[:n]
    doc = doc[n:]
    return doc, ret
}
