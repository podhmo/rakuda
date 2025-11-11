# Plan for `binding` Package

This document outlines the plan for creating a new `binding` package in the `rakuda` project. The package will provide a type-safe, reflect-free, and expression-oriented way to bind data from HTTP requests to Go structs, inspired by the `go-scan` example.

## 1. Core Concepts and Goals

- **Type Safety**: The binding functions will be generic, ensuring that the parsed values match the destination field's type.
- **Reflect-Free**: The implementation will not use reflection, which improves performance and avoids runtime panics associated with incorrect reflection usage.
- **Expression-Oriented**: The API will be functional and explicit, clearly showing the source, destination, and parsing logic for each field.
- **Extensibility**: A generic `Parser[T]` function type will allow users to define custom parsing logic for any type.

## 2. Package Structure and Components

The package will reside in a new `binding/` directory.

### Core Types

- `Binding`: A struct that holds the request context (`*http.Request`) and a function to extract path parameters.
  ```go
  type Binding struct {
      req       *http.Request
      pathValue func(string) string
  }
  ```
- `Source`: An enum-like `string` type to specify the data source.
  ```go
  type Source string

  const (
      Query  Source = "query"
      Header Source = "header"
      Cookie Source = "cookie"
      Path   Source = "path"
  )
  ```
- `Requirement`: A `bool` type to specify if a field is required or optional.
  ```go
  type Requirement bool

  const (
      Required Requirement = true
      Optional Requirement = false
  )
  ```
- `Parser[T any]`: A generic function type for parsing a string into a value of type `T`.
  ```go
  type Parser[T any] func(string) (T, error)
  ```

## 3. Implementation Steps

1.  **Create `binding/binding.go`**:
    - Define the core types: `Binding`, `Source`, `Requirement`, `Parser`.
    - Implement `New(req, pathValue) *Binding` constructor.
    - Implement internal lookup helpers:
        - `Lookup(source, key) (string, bool)`: Retrieves a single value from the specified source.
        - `valuesFromSource(source, key) ([]string, bool)`: Retrieves multiple values (e.g., from query parameters or comma-separated headers).

2.  **Implement Binding Functions**:
    - `One[T any](b *Binding, dest *T, source Source, key string, parse Parser[T], req Requirement) error`: Binds a single, non-pointer value.
    - `OnePtr[T any](b *Binding, dest **T, source Source, key string, parse Parser[T], req Requirement) error`: Binds a single, pointer value (for optional fields).
    - `Slice[T any](b *Binding, dest *[]T, source Source, key string, parse Parser[T], req Requirement) error`: Binds a slice of values.
    - `SlicePtr[T any](b *Binding, dest *[]*T, source Source, key string, parse Parser[T], req Requirement) error`: Binds a slice of pointer values.

3.  **Create `binding/binding_test.go`**:
    - Write comprehensive unit tests for all binding functions.
    - Test cases should cover:
        - All `Source` types (Query, Header, Cookie, Path).
        - Both `Requirement` types (Required, Optional).
        - Successful binding scenarios.
        - Missing value scenarios (for both required and optional).
        - Parsing failure scenarios.
        - Slice binding with multiple query parameters (e.g., `?id=1&id=2`).
        - Slice binding with comma-separated values (e.g., `X-Values: 1,2,3`).
        - Edge cases like empty values.

4.  **Create `binding/README.md`**:
    - Add a clear and concise `README.md` file within the `binding` directory.
    - Provide a simple but complete usage example showing how to create a `Binding` instance and use the `One` and `Slice` functions to bind data to a struct.

## 4. Future Enhancements (Not in this implementation)

After the initial implementation, the following features could be considered:

- **Pre-defined Parsers**: Provide a set of common parsers (e.g., for `int`, `bool`, `time.Time`) in a sub-package like `binding/parsers`.

Based on user feedback, the following features are explicitly **out of scope**:

- **Struct Tag-Based Binding**: This approach, which would likely require reflection, is not desired. The current expression-oriented API is preferred.
- **Explicit Default Values**: The current behavior, where an optional and missing value results in the Go default zero value for the field, is considered sufficient.

This plan provides a clear path to implementing a robust and user-friendly request binding package.
