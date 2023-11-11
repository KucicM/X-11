package common


type Token struct {
    Id int
    Runes []rune
}

type Document struct {
    Path string
    Title string
    Tokens []Token
    Url string
    Description string
}
