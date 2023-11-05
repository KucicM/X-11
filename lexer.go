package main

import (
	"unicode"
)

type lexer struct {
    content []rune
}

func (l *lexer) next() string {
    l.remove_left_pad()

    if len(l.content) == 0 {
        return ""
    }

    if unicode.IsDigit(l.content[0]) {
        return l.takeUntil(unicode.IsDigit)
    }

    if unicode.IsLetter(l.content[0]) {
        return l.takeUntil(unicode.IsLetter)
    }

    return l.take(1)
}

func (l *lexer) takeUntil(fn func(rune) bool) string {
    n := 0
    for n < len(l.content) && fn(l.content[n]) {
        n++
    }
    return l.take(n)
}

func (l *lexer) take(n int) string {
    for i := 0; i < n; i++ {
        l.content[i] = unicode.ToLower(l.content[i])
    }
    ret := string(l.content[:n])
    l.content = l.content[n:]
    return ret
}


func (l *lexer) remove_left_pad() {
    for len(l.content) > 0 && unicode.IsSpace(l.content[0]) {
        l.content = l.content[1:]
    }
}

func Tokenize(content string) []string {
    lexer := lexer{[]rune(content)}
    out := make([]string, 0)

    for {
        if token := lexer.next(); token != "" {
            out = append(out, token)
        } else {
            break
        }
    }

    return out
}

