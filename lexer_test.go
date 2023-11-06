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

func BenchmarkToNgramMediumSequences(b *testing.B) {
    sequence := `Title: Understanding Go (Golang) Memory Management: A Comprehensive Overview

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

Go's memory management system, powered by its automatic garbage collector, is a cornerstone of its simplicity and efficiency. By providing developers with a seamless experience and eliminating the complexities of manual memory management, Go empowers programmers to focus on building robust, scalable, and concurrent applications. As the programming landscape continues to evolve, Go's efficient memory management remains a key factor in its widespread adoption and success in the software development community.`
    for i := 0; i < b.N; i++ {
        _ = ToNgrams(sequence, 5)
    }
    b.ReportAllocs()
}

func BenchmarkTokenizingMediumSequences2(b *testing.B) {
    sequence := `Title: Understanding Go (Golang) Memory Management: A Comprehensive Overview

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

Go's memory management system, powered by its automatic garbage collector, is a cornerstone of its simplicity and efficiency. By providing developers with a seamless experience and eliminating the complexities of manual memory management, Go empowers programmers to focus on building robust, scalable, and concurrent applications. As the programming landscape continues to evolve, Go's efficient memory management remains a key factor in its widespread adoption and success in the software development community.`
    for i := 0; i < b.N; i++ {
        _ = Tokenize(sequence)
    }
    b.ReportAllocs()
}
