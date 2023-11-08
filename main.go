package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kucicm/X-11/pkg/build"
	"github.com/kucicm/X-11/pkg/server"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
    mode := flag.String("mode", "server", "options: server , build")
    flag.Parse()

    switch strings.ToLower(*mode) {
    case "server":
        startServer()
    case "build":
        startBuilding()
    }
}

func startServer() {
    log.Println("starting server...")

    file, err := os.ReadFile("server.json")
    if err != nil {
        log.Fatalf("ERROR: cannot open server.json, %s", err)
    }

    var cfg server.ServerCfg
    if err := json.Unmarshal(file, &cfg); err != nil {
        log.Fatalf("ERROR: cannot unmarshal server.json, %s", err)
    }

    server.StartServer(cfg)
}

func startBuilding() {
    log.Println("bulding...")
    defer func(start time.Time) {
        log.Printf("building done in %s", time.Since(start))
    }(time.Now())

    file, err := os.ReadFile("build.json")
    if err != nil {
        log.Fatalf("ERROR: cannot open build.json, %s", err)
    }

    var cfg build.BuildCfg
    if err := json.Unmarshal(file, &cfg); err != nil {
        log.Fatalf("ERROR: cannot unmarshal build.json, %s", err)
    }

    build.BuildIndices(cfg)
}

