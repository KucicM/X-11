package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func BuildIndex(root_dir_path, index_save_path string) (*Trie, *SearchIndex) {
    log.Println("building index...")
    start := time.Now()

    trie := NewTrie()
    //searchIdxBuilder := NewTfIdf()
    searchIdxBuilder, err := NewSearchIndexBuilder()
    if err != nil {
        log.Fatal(err)
    }
    filepath.WalkDir(root_dir_path, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if d.IsDir() {
            return nil
        }

        parts := strings.Split(path, ".")
        if parts[len(parts) - 1] != "txt" {
            return nil
        }

        log.Printf("Indexing %s %+v", path, d)

        bytes, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        // TODO maybe just use []byte
        tokens := Tokenize(string(bytes))
        searchIdxBuilder.AddDocument(path, tokens)

        //tfIdf.Add(path, ngmas)
        /*
        for _, ngram := range ToNgrams(string(bytes), 1, 1) {
            //trie.Insert(ngram)
            tfIdf.Insert(path, ngram)
        }
        */

        return nil
    })
    trie.PopulateCache(10)
    log.Printf("index building took %v", time.Since(start))

    saveIndex(trie, searchIdxBuilder, index_save_path)

    searchIndex, err := NewSeachIndex()
    if err != nil {
        log.Fatalln(err)
    }

    return trie, searchIndex
}

func saveIndex(trie *Trie, searchIdxBuilder *SearchIndexBuilder, path string) {
    log.Printf("saving index to %v", path)
    start := time.Now()

    trie.Save(path)
    searchIdxBuilder.Close()

    log.Printf("index saving took %v", time.Since(start))
}

func LoadIndex(src_path string) {

}
