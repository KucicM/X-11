package main

import (
	"fmt"
	"reflect"
	"testing"
)

const shortSequence = "foo"
const mediumSequence = "foo bar baz"
const longSequence = `ChatGPT:
Title: Understanding Go (Golang) Memory Management: A Comprehensive Overview

Introduction:

Go, also known as Golang, is a statically typed, compiled programming language designed for simplicity and efficiency. One of its notable features is its robust memory management system, which plays a crucial role in ensuring efficient resource utilization and high-performance computing. In this essay, we will delve into the fundamentals of Go's memory management, exploring its key concepts, advantages, and best practices.

Go's Memory Management:

Go employs a garbage collector, an automatic memory management system that automatically deallocates unused memory, preventing memory leaks and dangling pointers. This garbage collector allows developers to focus on writing code without worrying about manual memory management, making Go a reliable choice for scalable applications and concurrent programming.

Key Concepts:

    Automatic Garbage Collection: Go's garbage collector runs concurrently with the program, periodically identifying and releasing unused memory blocks, ensuring efficient memory utilization.

    Memory Allocation: Go provides built-in functions like make() and new() for allocating memory. The make() function is used for slices, maps, and channels, while the new() function allocates memory for a specific data type and returns a pointer.

    Stack and Heap: Go manages memory using both stack and heap. Variables with fixed sizes are stored in the stack, offering fast access, while dynamic data structures like slices and maps are stored in the heap, allowing for flexible memory allocation.

Advantages of Go's Memory Management:

    Prevents Memory Leaks: Go's garbage collector ensures that unused memory is promptly released, preventing memory leaks that can lead to system crashes and degraded performance.

    Simplifies Development: By automating memory management, Go simplifies the development process, allowing programmers to focus on writing efficient code rather than managing memory resources manually.

    Supports Concurrent Programming: Go's memory management system is designed to support concurrent programming, making it easier to develop highly concurrent applications without worrying about memory synchronization issues.

Best Practices for Memory Management in Go:

    Use Pointers Wisely: While Go abstracts pointers to make memory management easier, developers should use pointers judiciously, avoiding unnecessary indirections that can impact performance.

    Avoid Global Variables: Minimize the use of global variables, as they tend to stay in memory throughout the program's lifecycle. Instead, use local variables and pass them as arguments to functions.

    Profile and Optimize: Use Go's profiling tools to identify memory-intensive parts of the code. Once identified, optimize data structures and algorithms to reduce memory usage and improve overall performance.

Conclusion:

Go's memory management system, powered by its automatic garbage collector, is a cornerstone of its simplicity and efficiency. By providing developers with a seamless experience and eliminating the complexities of manual memory management, Go empowers programmers to focus on building robust, scalable, and concurrent applications. As the programming landscape continues to evolve, Go's efficient memory management remains a key factor in its widespread adoption and success in the software development community.
`

var testSequences = []string{shortSequence, mediumSequence, longSequence}

func BenchmarkTokenizing(b *testing.B) {
    for _, val := range testSequences {
        b.Run(fmt.Sprintf("str size %d", len(val)), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = Tokenize(val)
            }
            b.ReportAllocs()
        })
    }
}

func BenchmarkNgraming(b *testing.B) {
    for _, ngramSize := range []int{4, 16, 32, 64} {
        for _, val := range testSequences {
            b.Run(fmt.Sprintf("str size %d; ngmra size %d", len(val), ngramSize), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = ToNgrams(val, ngramSize, ngramSize)
            }
            b.ReportAllocs()
            })
        }
    }
}

func TestTokenizing(t *testing.T) {

    values := []struct {
        input string
        expected []Token
    }{
        {"Testing", toTokens("testing")},
        {"Testing something", toTokens("testing", "something")},
        {"Testing2", toTokens("testing", "2")},
        {"Testing 2", toTokens("testing", "2")},
        {"2Testing?", toTokens("2", "testing")},
        {"2 Testing!", toTokens("2", "testing")},
        {"let's test something.", toTokens("let", "s", "test", "something")},
        {"   spaces  are  something  else   ", toTokens("spaces", "are", "something", "else")},
    }

    for i, value := range values {
        t.Run(fmt.Sprintf("testing %d", i), func(t *testing.T) {
            tokens := Tokenize(value.input)
            if !reflect.DeepEqual(tokens, value.expected) {
                t.Errorf("expected %+v got %+v", value.expected, tokens)
            }
        })
    }
}

func TestNgrams(t *testing.T) {

    values := []struct {
        input string
        nMin int
        nMax int
        expected []Ngram
    }{
        {"Testing", 1, 1, toNgmas([]string{"testing"})},
        {"Testing2", 1, 2, toNgmas([]string{"testing"}, []string{"testing", "2"}, []string{"2"})},
        {"foo bar baz", 1, 2, toNgmas([]string{"foo"}, []string{"foo", "bar"}, []string{"bar"}, []string{"bar", "baz"}, []string{"baz"})},
        {"x x x x x", 2, 2, toNgmas([]string{"x", "x"}, []string{"x", "x"}, []string{"x", "x"}, []string{"x", "x"})},
        {"a b c d", 2, 3, toNgmas([]string{"a", "b"}, []string{"a", "b", "c"}, []string{"b", "c"}, []string{"b", "c", "d"}, []string{"c", "d"})},
    }

    for _, value := range values {
        t.Run(fmt.Sprintf("ngram (%d, %d): %s", value.nMin, value.nMax, value.input), func(t *testing.T) {
            ngrams := ToNgrams(value.input, value.nMin, value.nMax)
            if !reflect.DeepEqual(ngrams, value.expected) {
                t.Errorf("expected %+v got %+v", value.expected, ngrams)
            }
        })
    }
}

// unitily functions
func toNgmas(input ...[]string) []Ngram {
    out := make([]Ngram, 0, len(input))
    for _, in := range input {
        tokens := toTokens(in...)
        out = append(out, Ngram(tokens))
    }
    return out
}

func toTokens(input... string) []Token {
    out := make([]Token, 0, len(input))
    for _, in := range input {
        out = append(out, Token(in))
    }
    return out
}

