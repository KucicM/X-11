package main

import (
	"strings"
	"unicode"
)
type Ngram []Token

func (n Ngram) Size() int {
    total := 0
    for _, t := range n {
        total += len(t)
    }
    return total
}

type Token []rune
type source []rune

func ToNgrams(content string, max_chars int) []Ngram {
    out := make([]Ngram, 0)
    tokens := Tokenize(content)

    var buf Ngram = make([]Token, 0)
    for _, token := range tokens {
        buf = append(buf, token)
        cp := make([]Token, len(buf))
        copy(cp, buf)
        out = append(out, cp)
        for buf.Size() > max_chars {
            buf = buf[1:]
        }
    }
    return out
}

func Tokenize(content string) []Token {
    out := make([]Token, 0)

    src := source(content)
    for len(src) > 0 {
        var token Token
        src, token = src.next()
        if len(token) > 0 {
            out = append(out, token)
        }
    }

    return out
}

func UnTokenize(tokens []Token) string {
    strs := make([]string, 0, len(tokens))
    for _, token := range tokens {
        strs = append(strs, string(token))
    }
    return strings.Join(strs, " ")
}

func (src source) next() (source, Token) {
    for len(src) > 0 && unicode.IsSpace(src[0]) {
        src = src[1:]
    }

    if len(src) == 0 {
        return src, Token{}
    }

    if unicode.IsPunct(src[0]) {
        return src[1:], Token{}
    }

    if unicode.IsDigit(src[0]) {
        return src.takeUntil(unicode.IsDigit)
    }

    if unicode.IsLetter(src[0]) {
        return src.takeUntil(unicode.IsLetter)
    }

    return src.take(1)
}

func (src source)takeUntil(fn func(rune) bool) (source, Token) {
    n := 0
    for n < len(src) && fn(src[n]) {
        n++
    }
    return src.take(n)
}

func  (src source) take(n int) (source, Token) {
    for i := 0; i < n; i++ {
        src[i] = unicode.ToLower(src[i])
    }
    ret := src[:n]
    src = src[n:]
    return src, Token(ret)
}
