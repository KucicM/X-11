package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const rebuildIndex = true

func BuildIndex(root_dir_path, index_save_path string) (*Trie, *SearchIndex) {
    trie := NewTrie()
    if rebuildIndex {
        log.Println("building index...")
        start := time.Now()
        searchIdxBuilder := NewSearchIndexBuilder()

        _ = filepath.WalkDir(root_dir_path, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }

            if d.IsDir() {
                return nil
            }

            if !strings.HasSuffix(path, ".txt") {
                return nil
            }

            log.Printf("Indexing %s %+v", path, d)

            bytes, err := os.ReadFile(path)
            if err != nil {
                return err
            }

            // TODO maybe just use []byte
            //tokens := Tokenize(string(bytes))
            strTokens := StrTokenize(string(bytes))
            searchIdxBuilder.AddDocument(d.Name(), path, strTokens)

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
    }

    searchIndex := NewSeachIndex()

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
