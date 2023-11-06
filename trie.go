package main

import (
	"sort"
)

type Trie struct {
    root *trieNode
}

// todo store trie to disk maybe like key value
// also building trie should maybe use disk as well
func BuildTrie(words chan Token, topNcache int) *Trie {
    trie := &Trie{root: newTrieNode()}
    var node *trieNode

    for word := range words {
        node = trie.root
        for _, w := range word {
            node = node.next_or_create(w)
        }
        node.markFinal(word)
    }

    trie.populateCache(topNcache)
    return trie
}

func LoadTrie(path string) (*Trie, error) {
    return nil, nil
}

func (t *Trie) Save(path string) error {
    return nil
}

func (t *Trie) populateCache(n int) {
    t.root.populateCache(n)
}

func (t *Trie) finaAll(prefix Token) []Token {
    node := t.root
    for i := 0; i < len(prefix) && node != nil; i++ {
        node = node.next(prefix[i])
    }

    if node == nil {
        return []Token{}
    }
    return node.topN()
}

type tokenFreq struct {
    freq uint
    token Token
}

type trieNode struct {
    children map[rune]*trieNode
    word tokenFreq
    topNWords []*tokenFreq
}

func (n *trieNode) populateCache(topN int) []*tokenFreq {
    n.topNWords = make([]*tokenFreq, 0, len(n.children))

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
    return &trieNode{children: make(map[rune]*trieNode), word: tokenFreq{}}
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

func (n *trieNode) markFinal(token Token) {
    n.word.freq += 1
    n.word.token = token
}

func (n *trieNode) topN() []Token {
    ret := make([]Token, 0, len(n.topNWords))
    for _, w := range n.topNWords {
        ret = append(ret, w.token)
    }
    return ret
}
