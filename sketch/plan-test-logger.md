# Plan: Add t.Logf-based Logger to rakudatest

## Objective

To improve the debugging experience during tests, we will add a feature to `rakudatest` that injects a logger into the `http.Request` context. This logger will use `t.Logf` to record log messages, making them visible when a test fails.

This was achieved by implementing a custom `slog.Handler` that wraps a `*testing.T` instance and using `*slog.Logger` directly throughout the application, rather than introducing a custom logger interface.

## Tasks

### 1. Implement `slog` Handler for `*testing.T`

- **File**: `rakudatest/slog.go`
- **Purpose**: To create an adapter that allows `slog` to be used with `*testing.T`.
- **Details**:
    - A `THandler` struct was created that implements the `slog.Handler` interface.
    - The `Handle` method formats the log record and writes it to the `*testing.T` instance using `t.Logf()`.

### 2. Create Centralized Context Utilities

- **File**: `context.go`
- **Purpose**: To provide standardized, unexported functions for managing values in `context.Context`.
- **Details**:
    - `NewContextWithLogger` and `LoggerFromContext` were created to manage the `*slog.Logger` in the context.
    - `NewContextWithStatusCode` and `StatusCodeFromContext` were moved here from the `responder` package to centralize context management.

### 3. Update `rakudatest`

- **File**: `rakudatest/rakudatest.go`
- **Purpose**: To integrate the new logger into the test helper.
- **Details**:
    - The `Do` function was modified to accept a `*testing.T`.
    - Inside `Do`, an instance of the `THandler` is created and used to build a new `slog.Logger`.
    - This logger is injected into the request's context using `NewContextWithLogger`.

### 4. Add a Test Handler to Verify Logging

- **File**: `rakudatest/rakudatest_test.go`
- **Purpose**: To verify that the logger is correctly injected and used.
- **Details**:
    - A `spyHandler` was created that retrieves the logger from the request context and logs messages at various levels.
    - This handler is used in a new test case with `rakudatest.Do` to confirm the logger is injected and working as expected.
    - Existing tests for `rakudatest.Do` were restored and verified.

### 5. Refactor Application to Use `*slog.Logger`

- **Files**: `responder.go`, `lift.go`, `builder.go`, `rakudamiddleware/recover.go`, and all `_test.go` files.
- **Purpose**: To remove the custom `Logger` interface and use the standard `*slog.Logger` directly.
- **Details**:
    - All function signatures and variable types were updated from the custom interface to `*slog.Logger`.
    - The `responder_test.go` was updated to use a `testHandler` that implements `slog.Handler` for capturing log output during tests.

### 6. Update `TODO.md`

- A new section for this feature with all the tasks listed above was added.
