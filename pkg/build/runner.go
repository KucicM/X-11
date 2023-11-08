// entry point to build all indices (full text search index and autocomplete index)
package build

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kucicm/X-11/pkg/common"
)

type BuildCfg struct {
    SourceFolder string `json:"source-folder"`
    FtsCfg FullTextIndexCfg `json:"full-text-index"`
    TokenizerCfg common.TokenizerCfg `json:"tokenizer"`
}

func BuildIndices(cfg BuildCfg) {

    tokenizer := common.NewTokenizer(cfg.TokenizerCfg)
    defer tokenizer.Close()

    fts := newFullTextIndex(cfg.FtsCfg)
    defer fts.close()

    err := filepath.WalkDir(cfg.SourceFolder, func(path string, d fs.DirEntry, err error) error {
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

        tokens := tokenizer.Tokenize(bytes)
        fts.AddDocument(d.Name(), path, tokens)

        return nil
    })

    if err != nil {
        log.Fatalf("ERROR: walk failed %s", err)
    }
}
