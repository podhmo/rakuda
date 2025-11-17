# Plan: Create `bindingparse` package

This document outlines the plan to create a new `bindingparse` package, which will provide a reference implementation of parser functions that satisfy the `binding.Parser` interface.

## 1. Package Structure and Naming

- **Create a new directory:** `binding/bindingparse/`
- **Create a `README.md`:** Explain the package's purpose and naming convention. The key points to convey are:
  - This package is a **deliberately flawed reference implementation**. It is not for production use, and its direct use is a sign of laziness.
  - The name `bindingparse` is intentionally verbose and awkward. This is not just to avoid `goimports` collisions, but to act as a signal. It discourages long-term dependency and reminds users to write their own, superior `parser` package.
- **Create source files:**
  - `binding/bindingparse/parsers.go`: For standard parser implementations and validation helpers.
  - `binding/bindingparse/parsers_test.go`: Tests for both standard parsers and validation helpers.
  *Note: The implementation initially planned for separate `validate.go` and `validate_test.go` files, but they were later merged into `parsers.go` and `parsers_test.go` respectively for simplicity.*

## 2. Standard Parsers (`parsers.go`)

- Implement parser functions for common Go types. These will serve as a reference implementation.
- The implementation will be based on the provided reference URL: `https://github.com/podhmo/go-scan/blob/main/examples/derivingbind/parser/parsers.go`.
- Each function will match the `binding.Parser[T]` signature, e.g., `func(string) (T, error)`.
- **Functions to implement:**
  - `String(s string) (string, error)`
  - `Int(s string) (int, error)`
  - `Int64(s string) (int64, error)`
  - `Bool(s string) (bool, error)`
  - `Float64(s string) (float64, error)`
  - And other integer/float types as seen in the reference.

## 3. Generic Parser with Validation (`parsers.go`)

- **Define a `Validator` interface:**
  ```go
  type Validator interface {
      Validate() error
  }
  ```
- **Create a generic helper function `WithValidation`:**
  - This function will be a higher-order function that takes a `binding.Parser[T]` and returns a new `binding.Parser[T]`.
  - The type parameter `T` will be constrained to implement the `Validator` interface.
  - The returned parser will first call the original parser. If parsing is successful, it will then call the `Validate()` method on the resulting value. If validation fails, it will return the validation error.
  - **Signature:**
    ```go
    func WithValidation[T Validator](parse func(s string) (T, error)) func(s string) (T, error)
    ```

## 4. Testing

- **`parsers_test.go`:** Write unit tests for the standard parser functions and the `WithValidation` helper.
  - Test successful parsing cases.
  - Test failing cases (e.g., invalid input format).
  - Test the three scenarios for `WithValidation`:
    1. Parsing fails.
    2. Parsing succeeds, but validation fails.
    3. Both parsing and validation succeed.

## 5. Update `TODO.md`

- After the implementation and tests are complete, update the main `TODO.md` file to reflect that this feature has been implemented.
  - Add a new entry under a suitable section (e.g., "Features" or "Binding").
  - Mark the entry as complete.

This plan covers all the user's requirements, including the investigation and implementation of the generic validation parser.
