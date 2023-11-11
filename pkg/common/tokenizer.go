package common

import (
	"strings"
	"unicode"
)

func TokenizeBytes(content []byte) []string {
    return TokenizeStr(string(content))
}

func TokenizeStr(contnet string) []string {
    return TokenizeRunes([]rune(contnet))
}

func TokenizeRunes(content []rune) []string {
    terms := make([]string, 0)
    for len(content) > 0 {
        content = removeLeftPadding(content, func(r rune) bool {
            return unicode.IsSpace(r) || unicode.IsPunct(r)
        })

        if len(content) == 0 {
            break
        }

        var term []rune
        if unicode.IsDigit(content[0]) {
            content, term = takeUntil(content, unicode.IsDigit)
        } else if unicode.IsLetter(content[0]) {
            content, term = takeUntil(content, unicode.IsLetter)
        } else {
            content, term = take(content, 1)
        }
        if len(term) > 0 {
            terms = append(terms, string(term))
        }
    }
    return terms
}

func GramifyStr(content string, minN, maxN int) []string {
    terms := TokenizeStr(content)
    return gramify(terms, minN, maxN)
}

func gramify(terms []string, minN, maxN int) []string {
    if len(terms) < minN {
        return make([]string, 0)
    }

    ngmraTerms := make([]string, 0, len(terms))
    for l, h := 0, minN; h <= len(terms); {
        if (h - l) > maxN {
            l += 1
        }

        for ll := l; (h - ll) >= minN; ll++ {
            term := strings.Join(terms[ll:h], " ")
            if len(term) > 0 {
                ngmraTerms = append(ngmraTerms, term)
            }
        }

        if h <= len(terms) {
            h++
        }
    }
    return ngmraTerms
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
