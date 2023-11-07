package main

import "unicode"

/*
input is []byte (document) which is converted into Tokens.
Tokens have two parts:
- encoeded representation (int) and []rune

Tokenize -> produces individual tokens
Gramify -> produce n-gram tokens (multiple tokens are merged into single token)
*/

type Token struct {
    encoded int
    runes []rune
}

type document []rune
func (d document) len() int {
    return len(d)
}

func (d document) getNext() (document, []rune) {
    return nil, nil
}

func (d document) next() (document, []rune) {
    for len(d) > 0 && unicode.IsSpace(d[0]) {
        d = d[1:]
    }

    if len(d) == 0 {
        return d, nil
    }

    if unicode.IsPunct(d[0]) {
        return d[1:], nil
    }

    if unicode.IsDigit(d[0]) {
        return d.takeUntil(unicode.IsDigit)
    }

    if unicode.IsLetter(d[0]) {
        return d.takeUntil(unicode.IsLetter)
    }

    return d.take(1)
}

func (d document) takeUntil(fn func(rune) bool) (document, []rune) {
    n := 0
    for n < len(d) && fn(d[n]) {
        n++
    }
    return d.take(n)
}

func  (d document) take(n int) (document, []rune) {
    for i := 0; i < n; i++ {
        d[i] = unicode.ToLower(d[i])
    }
    ret := d[:n]
    d = d[n:]
    return d, ret
}

func Tokenize(content []byte) []Token {
    doc := document(string(content))
    tokens := make([]Token, 0)
    for doc.len() > 0 {
        var token []rune
        if doc, token = doc.getNext(); token != nil && len(token) > 0 {
            tokens = append(tokens, )
        }
    }
    return tokens
}

func Gramify(content []byte, minN, maxN int) []Token {
    tokens := Tokenize(content)

    if len(tokens) < minN {
        return make([]Token, 0)
    }

    out := make([]Token, 0, len(tokens))
    for l, h := 0, minN; h <= len(tokens); {
        if (h - l) > maxN {
            l += 1
        }

        for ll := l; (h - ll) >= minN; ll++ {
            out = append(out, mergeTokens(tokens[ll:h]))
        }

        if h <= len(tokens) {
            h++
        }
    }

    return out
}

func mergeTokens(tokens []Token) Token {
    total := 0
    for i := 0; i < len(tokens); i++ {
        total += len(tokens[i].runes) + 1
    }

    runes := make([]rune, total)
    for i := 0; i < len(tokens); i++ {
        runes = append(runes, tokens[i].runes...)
    }
    return // TODO
}

