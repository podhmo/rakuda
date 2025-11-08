## Design Document: The `responder` Package

**Author:** Gemini
**Date:** September 7, 2025
**Status:** Proposed

### 1. Abstract

This document proposes the design for a new Go package, `responder`, intended to simplify and standardize HTTP response generation in applications using the standard `net/http` library. It aims to eliminate boilerplate code by providing declarative, context-driven functions for common response types (e.g., JSON). Key state, such as the HTTP status code and a logger, is passed via the `http.Request`'s `context.Context`, promoting a clean, middleware-friendly architecture that is highly testable.

### 2. Background and Motivation

Writing HTTP handlers in Go using only the standard library is powerful but often involves repetitive and error-prone boilerplate for each handler:

1.  Setting the `Content-Type` header.
2.  Writing the HTTP status code.
3.  Encoding the response payload (e.g., marshaling a struct to JSON).
4.  Handling potential encoding errors.
5.  Logging errors consistently.

This repetition clutters handler logic, makes the code harder to read, and can lead to inconsistencies (e.g., forgetting to set a header).

The `responder` package aims to solve this by encapsulating this boilerplate into simple, reusable functions. By leveraging the `context.Context`, it decouples the response logic from the handler's business logic, allowing middleware to seamlessly modify response parameters like the status code or inject application-wide dependencies like a logger.

### 3. Goals

*   **Declarative API:** Handlers should be able to send a complete response with a single function call (e.g., `responder.JSON(w, r, data)`).
*   **Context-Driven:** All cross-cutting concerns (status code, logger) will be passed through the request `context`. This removes the need for a `Responder` struct and simplifies dependency management.
*   **High Testability:** The design must allow for easy unit testing of handlers. By using a logger interface and context injection, developers can provide mocks, such as a `*testing.T`-based logger, during tests.
*   **Minimalism:** The package should have a minimal and intuitive API, focusing on its core responsibility: writing responses.
*   **Safety:** The package should gracefully handle client disconnects by checking the request context for cancellation before writing the response.

### 4. Non-Goals

*   **Routing:** The package will not be a router or a web framework. It is designed to complement `net/http`'s `ServeMux` or any third-party router.
*   **Request Parsing/Validation:** The responsibility of parsing and validating incoming request data remains with the user's handler or dedicated middleware.
*   **Request Body Draining:** Draining and closing the request body to ensure HTTP Keep-Alive works correctly is considered an orthogonal concern, best handled by a dedicated middleware.

### 5. Proposed Design & API

The package will be purely functional, exporting functions and a single interface.

#### 5.1. `Logger` Interface

To decouple the package from any specific logging implementation (like `slog`), a minimal logger interface will be defined.

```go
package responder

import "context"

// Logger defines the minimal interface for a structured error logger.
// It is compatible with the slog.Logger and can be easily implemented
// by wrappers around other loggers or for testing purposes.
type Logger interface {
    ErrorContext(ctx context.Context, msg string, args ...any)
}
```

#### 5.2. Context Helper Functions

These functions provide a type-safe way to add and retrieve values from the request context.

```go
package responder

import (
    "context"
    "net/http"
)

// WithLogger returns a new request with the provided Logger stored in its context.
// This should typically be called once by a middleware at the top level.
func WithLogger(r *http.Request, logger Logger) *http.Request {
    // ... implementation using context.WithValue ...
}

// WithStatusCode returns a new request with the provided HTTP status code
// stored in its context. This can be called by any middleware or handler
// to set or override the status for the final response.
func WithStatusCode(r *http.Request, status int) *http.Request {
    // ... implementation using context.WithValue ...
}
```

Internal, unexported functions `getLogger(ctx)` and `getStatusCode(ctx)` will be used to retrieve these values, providing sane defaults (`nil` for logger, `http.StatusOK` for status).

#### 5.3. Responder Functions

The primary function for sending a JSON response.

```go
package responder

import "net/http"

// JSON marshals the 'data' payload to JSON and writes it to the response.
//
// It performs the following steps:
// 1. Checks if the request context has been canceled (e.g., client disconnected).
//    If so, it returns immediately to prevent "broken pipe" errors.
// 2. Retrieves the HTTP status code from the request context. If not set,
//    it defaults to http.StatusOK (200).
// 3. Sets the "Content-Type" header to "application/json; charset=utf-8".
// 4. Writes the HTTP status code to the response header.
// 5. If data is not nil, it encodes the data to the response writer.
// 6. If encoding fails, it retrieves the Logger from the context. If a logger
//    exists, it logs the error with contextual information.
func JSON(w http.ResponseWriter, req *http.Request, data any) {
    // ... implementation ...
}
```

### 6. Usage Example

#### 6.1. Application Usage

```go
// main.go
package main

import (
    "log/slog"
    "net/http"
    "os"
    "responder" // This package
)

// LoggerMiddleware injects the application-wide logger into each request context.
func LoggerMiddleware(logger responder.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            next.ServeHTTP(w, responder.WithLogger(r, logger))
        })
    }
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := findUserByID(r.Context(), r.URL.Query().Get("id"))

    if err == errNotFound {
        r = responder.WithStatusCode(r, http.StatusNotFound)
        responder.JSON(w, r, map[string]string{"error": "user not found"})
        return
    }
    if err != nil {
        r = responder.WithStatusCode(r, http.StatusInternalServerError)
        // The error will be logged automatically by the JSON function if encoding fails,
        // but here we log the application error itself.
        logger, _ := r.Context().Value(loggerKey).(responder.Logger)
        if logger != nil {
            logger.ErrorContext(r.Context(), "failed to find user", "error", err)
        }
        responder.JSON(w, r, map[string]string{"error": "internal server error"})
        return
    }

    // On success, the default status 200 OK is used.
    responder.JSON(w, r, user)
}

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    mux := http.NewServeMux()
    mux.HandleFunc("/api/user", getUserHandler)

    // Apply middleware
    var handler http.Handler = mux
    handler = LoggerMiddleware(logger)(handler)

    http.ListenAndServe(":8080", handler)
}```

#### 6.2. Testing Usage

```go
// handler_test.go
package main

import (
    "context"
    "net/http"
    "net/http/httptest"
    "responder"
    "testing"
)

// testLogger adapts *testing.T to the responder.Logger interface.
type testLogger struct {
    t *testing.T
}

func (l *testLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
    l.t.Helper()
    l.t.Errorf("Error Log: %s | Args: %v", msg, args)
}

func TestGetUserHandler_NotFound(t *testing.T) {
    // Arrange
    req := httptest.NewRequest(http.MethodGet, "/api/user?id=nonexistent", nil)
    rr := httptest.NewRecorder()

    // Inject test logger
    logger := &testLogger{t: t}
    req = responder.WithLogger(req, logger)

    // Act
    getUserHandler(rr, req)

    // Assert
    if status := rr.Code; status != http.StatusNotFound {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
    }

    expectedBody := `{"error":"user not found"}` + "\n"
    if rr.Body.String() != expectedBody {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
    }
}
```

### 7. Extensibility

This design is easily extensible. Other responder functions can be added following the same pattern:

*   `responder.HTML(w, r, templateName, data)`
*   `responder.XML(w, r, data)`
*   `responder.File(w, r, filePath)`
*   `responder.NoContent(w, r)`

Each new function would respect the status code and logger set in the context, ensuring a consistent and predictable API.
