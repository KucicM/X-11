package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func TestSearch(t *testing.T) {
    index := NewSeachIndex()
    tokens := StrTokenize("In this example, specifies the maximum number of open connections in the pool. When you perform database operations concurrently from multiple goroutines, each goroutine will obtain a connection from the pool. If all connections are in use, additional goroutines will wait until a connection becomes available. Remember to properly handle errors and ensure that you close the connections after using them to avoid connection leaks. Additionally, consider using a proper error handling mechanism and potentially a connection pool library for more advanced use cases.")

    for i := 1; i <= 64; i *= 2 {
        start := time.Now()
        _, _ = index.Search(tokens[:i], 10)
        fmt.Printf("%d, %v\n", i, time.Since(start))
    }
}

func BenchmarkSearch(b *testing.B) {
    index := NewSeachIndex()

    var wg sync.WaitGroup
    tokens := StrTokenize("clear")
    for i := 0; i < b.N; i++ {
        wg.Add(1)
        go func() {
            _, _ = index.Search(tokens, 10)
            wg.Done()
        }()
    }
    wg.Wait()
    b.ReportAllocs()
}

func BenchmarkOpenClose(b *testing.B) {
    db, _ := sqlx.Open("sqlite3", "test.db?mode=ro")
    for i := 0; i < b.N; i++ {
        db.Query("SELECT 1;")
    }
    db.Close()
    b.ReportAllocs()
}
