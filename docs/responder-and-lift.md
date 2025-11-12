# Design Document: `responder` and `lift`

- **Author:** Gemini
- **Status:** Proposed
- **Date:** 2025-11-11

## 1. Abstract

This document outlines the design for two key components that work together to manage the boundary between HTTP transport and application business logic: the `responder` package and the `lift` function. The `responder` is a utility focused on the final step of writing a response (e.g., JSON, redirects). The `lift` function is a generic adapter that converts a Go function ("Action") into a standard `http.Handler`, using the `responder` to handle the output. This separation of concerns simplifies handler logic, centralizes response-writing, and improves type safety.

## 2. Terminology

To ensure clarity, we define the following terms:

- **Handler:** The standard Go `http.Handler` or `http.HandlerFunc`. Its responsibility is to interpret an `http.Request` and write to an `http.ResponseWriter`. Handlers are concerned with HTTP-level details.
- **Action:** A Go function with the signature `func(r *http.Request) (O, error)`, where `O` is a generic output type. Its responsibility is to execute business logic and return data or an error. An Action should be completely unaware of the `http.ResponseWriter`.
- **Responder:** A utility that takes the output from a Handler (or a `lift`-wrapped Action) and writes it to the `http.ResponseWriter`. It handles JSON encoding, setting headers, and writing status codes.

## 3. The `responder` Component

### 3.1. Motivation

Writing HTTP responses in Go involves boilerplate for setting headers, encoding data, and handling errors. The `responder` encapsulates this logic into a simple, reusable component.

### 3.2. Design

The `Responder` is a struct that holds dependencies like a logger. It has methods for writing different kinds of responses:

- `JSON(w, r, statusCode, data)`: Encodes `data` as JSON.
- `Error(w, r, statusCode, err)`: Writes a structured JSON error. It handles special cases like `*binding.ValidationErrors` and scrubs internal error details for 5xx status codes.
- `Redirect(w, r, url, statusCode)`: Performs an HTTP redirect.
- `NoContent(w, r, statusCode)`

The `Responder` is intended to be instantiated once and shared across the application.

```go
// The Responder is instantiated once.
var res = responder.New(&responder.Config{...})

// It's used directly inside a standard http.Handler.
func myHandler(w http.ResponseWriter, r *http.Request) {
    // ... logic ...
    res.JSON(w, r, http.StatusOK, myData)
}
```

For most applications, a single `Responder` instance is sufficient. It is designed to be goroutine-safe and can be defined as a global variable. Future enhancements may include making the `Responder` more configurable (e.g., to pass options to an underlying JSON encoder like `encoding/json/v2`), but this will be done without requiring complex configuration files.

## 4. The `lift` Function

### 4.1. Motivation

While the `responder` simplifies the final step of writing a response, the application logic inside a handler can still be complex, mixing business logic with HTTP concerns. `lift` solves this by adapting a pure business logic function (an "Action") to the `http.Handler` interface.

### 4.2. Design

`lift` is a generic higher-order function with the following signature:

```go
func Lift[O any](responder *Responder, action func(*http.Request) (O, error)) http.Handler
```

It acts as a bridge:
1. It returns an `http.HandlerFunc`.
2. Inside this handler, it calls the `action` function.
3. It inspects the `(O, error)` return values from the `action`.
4. It uses the `responder` to write the final HTTP response based on the outcome.

This separates the `action`'s business logic from the `responder`'s HTTP-writing logic.

### 4.3. Controlling the Response from an Action

An Action cannot directly write the status code or headers. Instead, it signals its intended outcome through its return values.

- **Success:**
  - `return data, nil`: The `responder` writes a `200 OK` with a JSON body of `data`.
  - If `data` implements `StatusCode() int`, that code is used instead of 200.

- **Failure:**
  - `return nil, err`: The `responder` writes an error response.
  - To control the status code, `err` can be wrapped (e.g., `return nil, rakuda.NewAPIError(http.StatusNotFound, err)`).
  - If `err` is a `*binding.ValidationErrors`, the status code is automatically `400 Bad Request`.
  - Otherwise, the default is `500 Internal Server Error`.

- **Redirects:**
  - `return nil, &rakuda.RedirectError{URL: "/login", Code: http.StatusFound}`: Returning this special error type signals to `lift` that it should use `responder.Redirect`.

- **Empty Responses:**
  - `return nil, nil` with a slice type `O`: `200 OK` with an empty JSON array `[]`.
  - `return nil, nil` with a map type `O`: `200 OK` with an empty JSON object `{}`.
  - `return nil, nil` with any other nillable type `O` (e.g., a pointer): `204 No Content`.


### 4.4. Limitations and Design Trade-offs

The design of `lift` is influenced by some of Go's language features.

- **Methods Cannot Have Type Parameters:** Ideally, `lift` might be a method on the `Responder` struct, like `responder.Lift(...)`. However, Go methods cannot be generic. This is why `Lift` is a standalone function that takes the `Responder` as an argument.

- **No `(I, O)` Input and Output Generics:**
  A more advanced `lift` function might have a signature like `func[I, O any](action func(r *http.Request, input I) (O, error))`. This would automatically handle request body parsing and binding for the input `I`. While this is a potential future enhancement, the current design prioritizes simplicity by focusing only on the output `O`. The input binding is handled separately by the `binding` package.

## 5. Summary: `handler` vs. `action`

- **Use a standard `Handler` (`http.HandlerFunc`) when:**
  - You need fine-grained control over the `http.ResponseWriter` (e.g., streaming, setting cookies, custom headers).
  - The logic is simple and primarily concerned with HTTP.
  - **Inside a Handler, you will call `responder` methods directly.**

- **Use an `Action` (wrapped with `lift`) when:**
  - The logic is primarily business logic that can be decoupled from HTTP.
  - You want to return data or an error and let `lift` handle the HTTP response details.
  - The function is easier to test without needing an `http.ResponseWriter`.
  - **Inside an Action, you *never* call the `responder`. You signal the outcome via return values.**

This clear separation allows developers to choose the right tool for the job, keeping code clean, testable, and maintainable.
