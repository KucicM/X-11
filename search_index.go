package main

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func BuildIndex(root_dir_path, index_save_path string) (*Trie, *TfIdf) {
    log.Println("building index...")
    start := time.Now()

    trie := NewTrie()
    tfIdf := NewTfIdf()
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

        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)

        for scanner.Scan() {
            for _, ngram := range ToNgrams(scanner.Text(), 1, 1) {
                //trie.Insert(ngram)
                tfIdf.Insert(path, ngram)
            }
        }

        return nil
    })
    trie.PopulateCache(10)
    log.Printf("index building took %v", time.Since(start))

    saveIndex(trie, tfIdf, index_save_path)

    return trie, tfIdf
}

func saveIndex(trie *Trie, tfIdf *TfIdf, path string) {
    log.Printf("saving index to %v", path)
    start := time.Now()

    trie.Save(path)
    tfIdf.Save(path)

    log.Printf("index saving took %v", time.Since(start))
}

func LoadIndex(src_path string) {

}
