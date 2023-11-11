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
}

type InputData struct {
    Title string `json:"title"`
    Text string `json:"text"`
    Url string `json:"url"`
    Description string `json:"description"`
}

type Document struct {
    Path string
    Title string
    Tokens []string
    Url string
    Description string
}

func BuildIndices(cfg BuildCfg) {
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

        titleTokens := common.GramifyStr(data.Title, 1, 1)
        textTokens := common.GramifyStr(data.Text, 1, 1)
        tokens := make([]string, 0, len(titleTokens) + len(textTokens))
        tokens = append(tokens, titleTokens...)
        tokens = append(tokens, textTokens...)

        doc := Document{
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
