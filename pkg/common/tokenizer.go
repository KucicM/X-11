package common

import (
	"unicode"
)


func Tokenize(contnet string) []string {
    runes := []rune(contnet)
    terms := make([]string, 0)
    for len(runes) > 0 {
        runes = removeLeftPadding(runes, func(r rune) bool {
            return unicode.IsSpace(r) || unicode.IsPunct(r)
        })

        if len(runes) == 0 {
            break
        }

        var term []rune
        if unicode.IsDigit(runes[0]) {
            runes, term = takeUntil(runes, unicode.IsDigit)
        } else if unicode.IsLetter(runes[0]) {
            runes, term = takeUntil(runes, unicode.IsLetter)
        } else {
            runes, term = take(runes, 1)
        }
        if len(term) > 0 {
            terms = append(terms, string(term))
        }
    }
    return terms
}

func removeLeftPadding(doc []rune, toRemove func(rune) bool) []rune {
    for len(doc) > 0 && toRemove(doc[0]) {
        doc = doc[1:]
    }
    return doc
}

func takeUntil(doc []rune, fn func(rune) bool) ([]rune, []rune) {
    n := 0
    for n < len(doc) && fn(doc[n]) {
        n++
    }
    return take(doc, n)
}

func take(doc []rune, n int) ([]rune, []rune) {
    for i := 0; i < n; i++ {
        doc[i] = unicode.ToLower(doc[i])
    }
    ret := doc[:n]
    doc = doc[n:]
    return doc, ret
}
