// entry point to build full text search index
package build

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kucicm/X-11/pkg/common"
	"github.com/kucicm/X-11/pkg/server"
)

type BuildCfg struct {
    SourceFolder string `json:"source-folder"`
    FtsCfg server.FullTextSearchCfg `json:"full-text-search"`
}

type InputData struct {
    Title string `json:"title"`
    Text string `json:"text"`
    Url string `json:"url"`
}

func BuildIndices(cfg BuildCfg) {
    fts := server.NewFullTextSearch(cfg.FtsCfg)
    defer fts.FinishIndexing()

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

        titleTokens := common.Tokenize(data.Title)
        textTokens := common.Tokenize(data.Text)
        tokens := make([]string, 0, len(titleTokens) + len(textTokens))
        tokens = append(tokens, titleTokens...)
        tokens = append(tokens, textTokens...)

        doc := server.Document{
            Title: data.Title,
            Tokens: tokens,
            Url: data.Url,
        }
        fts.AddDocument(doc)
        return nil
    })

    if err != nil {
        log.Fatalf("ERROR: walk failed %s", err)
    }
}
