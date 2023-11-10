package server

import (
	"github.com/kucicm/X-11/pkg/common"
)

type SearchIndexResult struct {
    Rank float64 `db:"rank"`
    Title string `db:"file_name"`
    Description string `db:"file_name"`
}

type SearchIndexCfg struct {
    Tokenizer common.TokenizerCfg `json:"tokenizer"`
    FullTextIndex FullTextIndexCfg `json:"full-text-index"`
}

type SearchIndex struct {
    tokenizer *common.Tokenizer
    index FullTextIndex
}

func NewSearchIndex(cfg SearchIndexCfg) *SearchIndex {
    return &SearchIndex{
        tokenizer: common.NewTokenizer(cfg.Tokenizer),
        index: NewFullTextIndex(cfg.FullTextIndex),
    }
}

func (i *SearchIndex) Search(query string, maxResults int) ([]SearchIndexResult, error) {
    ret := make([]SearchIndexResult, 0, maxResults)
    tokens := i.tokenizer.TokenizeStr(query)
    if len(tokens) == 0 {
        return ret, nil
    }

    return i.index.Search(tokens, maxResults)
}
