package pkg

/*
import (
	"log"
	"sort"
	"time"
)

type Prefix []Token

func createPrefix(tokens []Token) Prefix {
    prefix := make([]Token, 0, 2 * len(tokens))
    prefix = append(prefix, tokens[0])
    for i := 1; i < len(tokens); i++ {
        prefix = append(prefix, []rune(" "))
        prefix = append(prefix, tokens[i])
    }
    return prefix
}

type Trie struct {
    root *trieNode
}

// todo store trie to disk maybe like key value
// also building trie should maybe use disk as well
func NewTrie() *Trie {
    return &Trie{root: newTrieNode()}
}

func (t *Trie) Insert(ngram Ngram) {
    if len(ngram) == 0 {
        return 
    }
    prefix := createPrefix(ngram)
    node := t.root
    for _, token := range prefix {
        for _, t := range token {
            node = node.next_or_create(t)
        }
    }
    node.markFinal(prefix)
}

func LoadTrie(path string) (*Trie, error) {
    return nil, nil
}

func (t *Trie) Save(path string) error {
    return nil
}

func (t *Trie) PopulateCache(n int) {
    start := time.Now()
    log.Println("INFO: populating trie cache")
    t.root.populateCache(n)
    log.Printf("INFO: trie cache population took %v", time.Since(start))
}

func (t *Trie) Search(tokens []Token) []Prefix {
    prefix := createPrefix(tokens)
    node := t.root
    for i := 0; i < len(prefix) && node != nil; i++ {
        for j := 0; j < len(prefix[i]) && node != nil; j++ {
            node = node.next(prefix[i][j])
        }
    }

    if node == nil {
        return []Prefix{}
    }

    return node.topN()
}

type prefixFreq struct {
    freq uint
    prefix Prefix
}

type trieNode struct {
    children map[rune]*trieNode
    word prefixFreq
    topNWords []*prefixFreq
}

func (n *trieNode) populateCache(topN int) []*prefixFreq {
    n.topNWords = make([]*prefixFreq, 0, len(n.children))

    if len(n.children) == 0 && n.word.freq == 0 {
        return n.topNWords
    }

    n.topNWords = append(n.topNWords, &n.word)
    if len(n.children) == 0 {
        return n.topNWords
    }

    for _, child := range n.children {
        n.topNWords = append(n.topNWords, child.populateCache(topN)...)
    }

    sort.Slice(n.topNWords, func(i, j int) bool {
        return n.topNWords[i].freq > n.topNWords[j].freq
    })

    n.topNWords = n.topNWords[:min(topN, len(n.topNWords))]
    return n.topNWords
}

func newTrieNode() *trieNode {
    return &trieNode{children: make(map[rune]*trieNode), word: prefixFreq{}}
}

func (n *trieNode) next_or_create(b rune) *trieNode {
    if val, ok := n.children[b]; ok {
        return val
    }
    val := newTrieNode()
    n.children[b] = val
    return val
}

func (n *trieNode) next(b rune) *trieNode {
    return n.children[b]
}

func (n *trieNode) markFinal(prefix Prefix) {
    n.word.freq += 1
    n.word.prefix = prefix
}

func (n *trieNode) topN() []Prefix {
    ret := make([]Prefix, 0, len(n.topNWords))
    for _, w := range n.topNWords {
        ret = append(ret, w.prefix)
    }
    return ret
}
*/
