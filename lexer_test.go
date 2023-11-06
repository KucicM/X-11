package main

import "testing"

func BenchmarkTokenizingShortSequences(b *testing.B) {
    sequences := []string{"foo", "bar"}
    for i := 0; i < b.N; i++ {
        for _, sequence := range sequences {
            _ = Tokenize(sequence)
        }
    }
}

func BenchmarkTokenizingMediumSequences(b *testing.B) {
    sequences := []string{"foo bar baz", "foo bar baz"}
    for i := 0; i < b.N; i++ {
        for _, sequence := range sequences {
            _ = Tokenize(sequence)
        }
    }
}
