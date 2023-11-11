// entry point to build all indices (full text search index and autocomplete index)
package build

import (
	"encoding/json"
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

type InputData struct {
    Title string `json:"title"`
    Text string `json:"text"`
    Url string `json:"url"`
    Description string `json:"description"`
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

        if !strings.HasSuffix(path, ".json") {
            return nil
        }

        log.Printf("Indexing %s %+v", path, d)

        bytes, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        var data InputData
        if err := json.Unmarshal(bytes, &data); err != nil {
            return err
        }

        titleTokens := tokenizer.Tokenize([]byte(data.Title))
        textTokens := tokenizer.Gramify([]byte(data.Text), 1, 1)
        tokens := make([]common.Token, 0, len(titleTokens) + len(textTokens))
        tokens = append(tokens, titleTokens...)
        tokens = append(tokens, textTokens...)

        doc := common.Document{
            Path: path,
            Title: data.Title,
            Tokens: tokens,
            Url: data.Url,
            Description: data.Description,
        }
        fts.AddDocument(doc)

        return nil
    })

    if err != nil {
        log.Fatalf("ERROR: walk failed %s", err)
    }
}
