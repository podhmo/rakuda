# Design Document: `binding` Package

- **Author:** Gemini
- **Status:** Proposed
- **Date:** 2025-11-11

## 1. Abstract

This document details the design of the `binding` package, a type-safe, reflect-free, and expression-oriented library for binding data from HTTP requests to Go structs. The package is designed to provide a secure and predictable data-binding experience by eschewing magic (like struct tags and reflection) in favor of explicit, composable functions. It leverages Go generics to achieve strong type safety without sacrificing developer ergonomics. A key design principle is its decoupling from `net/http`, making it a more general-purpose binding utility.

## 2. Goals and Motivation

Data binding is a common source of bugs and security vulnerabilities in web applications. Many Go libraries rely on struct tags and `reflect`, which can lead to runtime panics, silent failures, and a disconnect between the struct definition and the binding logic.

The `binding` package is motivated by the following goals:

- **Type Safety:** The binding process should be fully checked by the Go compiler. An attempt to bind a string to an `int` field using an incorrect parser should result in a compile-time error.
- **Explicit and Composable:** Binding logic should be clear, explicit code, not hidden in struct tags. This makes it easy to read, debug, and compose complex validation rules. Binding operations should be simple expressions that can be combined.
- **No Reflection, No Magic:** The package must not use the `reflect` package for its core binding operations. This avoids runtime panics, improves performance, and makes the code easier for developers to reason about.
- **`net/http` Agnostic (mostly):** While the primary use case is HTTP requests, the core logic should be independent of `net/http`. The package's functions should operate on simple string inputs, making them testable and usable in other contexts. The link to `net/http` is contained within a thin `Binding` struct, which acts as an adapter.
- **Structured Error Reporting:** Errors should be structured and informative, clearly indicating the source, key, and value that caused the failure.

## 3. Proposed Architecture

### 3.1. Expression-Oriented API

The core design principle is that each binding operation is an "expression"—a function call that takes a destination, a source, and a parser.

**Example:**

```go
type MyInput struct {
    ID    int
    Sort  *string
    Tags  []string
}

func action(r *http.Request) (*MyInput, error) {
    b := binding.New(r, r.PathValue) // Adapter for http.Request
    var input MyInput

    // Each line is a self-contained binding expression.
    err := binding.Join(
        binding.One(b, &input.ID, binding.Path, "id", strconv.Atoi, binding.Required),
        binding.OnePtr(b, &input.Sort, binding.Query, "sort", binding.ToString, binding.Optional),
        binding.Slice(b, &input.Tags, binding.Query, "tags", binding.ToString, binding.Optional),
    )

    if err != nil {
        return nil, err
    }
    return &input, nil
}
```

This approach makes the binding logic self-documenting and easy to modify. All binding rules are co-located and visible, rather than scattered across a struct definition as tags.

### 3.2. Leveraging Generics for Type Safety

Go generics are central to the package's design. The primary binding functions, `One`, `OnePtr`, `Slice`, and `SlicePtr`, are generic.

```go
func One[T any](b *Binding, dest *T, source Source, key string, parse Parser[T], req Requirement) error
```

- **`T any`**: Represents the target type (e.g., `int`, `string`, `time.Time`).
- **`dest *T`**: The destination variable must be a pointer to `T`, ensuring the function can modify it.
- **`parse Parser[T]`**: The `Parser` function is also generic (`func(string) (T, error)`). This is the key to type safety. The compiler guarantees that the `Parser`'s return type `T` matches the destination's type `*T`. A mismatch (e.g., passing `strconv.Atoi` for a `string` field) will cause a compile-time error.

### 3.3. Explicit Parsers, No Coercion

The package does not perform any automatic type coercion. The developer must always provide an explicit `Parser` function. This eliminates ambiguity and forces conscious decisions about how a string from a request should be converted into a Go type. For simple cases, a `binding.ToString` parser can be used, but for numeric types, functions like `strconv.Atoi` are used directly.

### 3.4. Decoupling from `net/http`

The core binding functions (`One`, `Slice`, etc.) do not directly interact with `*http.Request`. Instead, they operate through a `Binding` struct.

```go
type Binding struct {
    req       *http.Request
    pathValue func(string) string
}

func New(req *http.Request, pathValue func(string) string) *Binding
```

The `Binding` struct acts as an **adapter**. It provides a `Lookup` method that knows how to extract a string value from different parts of a request (query, header, path, etc.). This design isolates the `net/http`-specific logic. In theory, one could create a `Binding` adapter for a different context, like a CLI application, and reuse the same core binding functions.

### 3.5. Structured Errors

The package defines two primary error types:

- **`binding.Error`**: Represents a single error, containing the source, key, invalid value, and underlying error.
- **`binding.ValidationErrors`**: A collection of `*binding.Error` instances.

The `binding.Join` function is a helper that aggregates multiple errors into a single `*ValidationErrors` instance. This error type implements a `StatusCode() int` method returning `400 Bad Request`, allowing it to integrate seamlessly with the `lift` handler, which automatically uses this status code for the HTTP response.

## 4. Rejected Alternatives

### 4.1. Struct Tags and Reflection

- **Description:** Use struct tags (e.g., `` `query:"id"` ``) and the `reflect` package to automatically populate a struct. This is the most common approach in the Go ecosystem.
- **Reason for Rejection:** This approach violates the core goals of the project.
    - **It's not type-safe:** Type mismatches are runtime errors, not compile-time errors.
    - **It's "magic":** The binding logic is hidden and inflexible. Custom validation requires more tags or hooks, increasing complexity.
    - **It's brittle:** It's prone to runtime panics if used incorrectly (e.g., non-pointer struct).

### 4.2. Fluent, Method-Chaining API

- **Description:** A fluent API could look like this: `binding.New(r).FromQuery("id").Required().AsInt(&input.ID)`.
- **Reason for Rejection:** While potentially elegant, this often requires more complex state management within the binding object. The chosen functional approach (`One(...)`, `Slice(...)`) is more aligned with Go's idiomatic, straightforward style. Each call is a self-contained, stateless expression, which is simpler to reason about and compose.
