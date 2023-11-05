package main

import (
	"sort"
)

type Trie struct {
    root *trieNode
}

func BuildTrie(words chan string, topNcache int) *Trie {
    trie := &Trie{root: newTrieNode()}
    var node *trieNode

    for word := range words {
        node = trie.root
        for _, w := range []byte(word) {
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

func (t *Trie) finaAll(prefix string) []string {
    node := t.root
    for i := 0; i < len(prefix) && node != nil; i++ {
        node = node.next(prefix[i])
    }

    if node == nil {
        return []string{}
    }
    return node.topN()
}

type wordFreq struct {
    freq uint
    word string
}

type trieNode struct {
    children map[byte]*trieNode
    word wordFreq
    topNWords []*wordFreq
}

func (n *trieNode) populateCache(topN int) []*wordFreq {
    n.topNWords = make([]*wordFreq, 0, len(n.children))

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
    return &trieNode{children: make(map[byte]*trieNode), word: wordFreq{}}
}

func (n *trieNode) next_or_create(b byte) *trieNode {
    if val, ok := n.children[b]; ok {
        return val
    }
    val := newTrieNode()
    n.children[b] = val
    return val
}

func (n *trieNode) next(b byte) *trieNode {
    return n.children[b]
}

func (n *trieNode) markFinal(word string) {
    n.word.freq += 1
    n.word.word = word
}

func (n *trieNode) topN() []string {
    ret := make([]string, 0, len(n.topNWords))
    for _, w := range n.topNWords {
        ret = append(ret, w.word)
    }
    return ret
}
