package main

import (
	"unicode"
)

type Token []rune
type source []rune

func Tokenize(content string) []Token {
    out := make([]Token, 0)

    src := source(content)
    for len(src) > 0 {
        var token Token
        src, token = src.next()
        out = append(out, token)
    }

    return out
}

func (src source) next() (source, Token) {
    for len(src) > 0 && unicode.IsSpace(src[0]) {
        src = src[1:]
    }

    if len(src) == 0 {
        return src, Token{}
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
