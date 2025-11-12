# TODO

> **Note on updating this file:**
> - Do not move individual tasks to the "Implemented" section.
> - A whole feature section (e.g., "Core Router Implementation") should only be moved to "Implemented" when all of its sub-tasks are complete.
> - For partially completed features, use checkboxes (`[x]` for complete, `[-]` for partially complete). A feature is considered partially complete if it has been implemented but has associated tests that are currently disabled.
> - For partially completed features, use checkboxes (`[x]`) to mark completed sub-tasks.

This file tracks implemented features and immediate, concrete tasks.

For more ambitious, long-term features, see [docs/router-design.md](./docs/router-design.md).

## Implemented

- **Project Structure**: Initial repository setup with LICENSE, README.md, and AGENTS.md.
- **Documentation**: Design document outlining the core architecture and philosophy ([docs/router-design.md](./docs/router-design.md)).
- **Responder Package**:
  - [x] Create `responder` package and basic implementation.
  - [x] Refactor: Integrate `responder` package into the main `rakuda` package.
    - [x] Rename `rakuda.go` to `builder.go`.
    - [x] Rename `rakuda_test.go` to `builder_test.go`.
    - [x] Move `responder` package contents to top-level `responder.go`.
    - [x] Refactor `JSON` function into a method on a `Responder` struct.
    - [x] Update tests to reflect the `Responder` struct changes.
    - [x] Remove the now-empty `responder` directory.
- **Lift Function**:
    - [x] Implement `lift` function to convert `func(req) (data, error)` into `http.Handler`.
    - [x] Support custom status codes via `APIError` struct with a `StatusCode()` method.
    - [x] Handle `(nil, nil)` returns: `204` for pointers, `[]` for slices, `{}` for maps.
- **rakudatest Test Helper**: Implemented a convenient test helper package for `http.Handler` testing. ([sketch/plan-rakudatest.md](./sketch/plan-rakudatest.md))
    - [x] Includes a generic `Do[T any]` function that integrates request execution, status code validation, optional response assertions, and JSON decoding into a single call.
    - [x] The `Do` function logs the full response body on a status code mismatch, which was the primary requirement.
- **Binding Package**: Implemented a type-safe, reflect-free request data binding package. ([sketch/plan-binding.md](./sketch/plan-binding.md))
    - [x] **Create `binding/binding.go`**:
        - [x] Define the core types: `Binding`, `Source`, `Requirement`, `Parser`.
        - [x] Implement `New(req, pathValue) *Binding` constructor.
        - [x] Implement internal lookup helpers.
    - [x] **Implement Binding Functions**:
        - [x] `One[T any](...) error`
        - [x] `OnePtr[T any](...) error`
        - [x] `Slice[T any](...) error`
        - [x] `SlicePtr[T any](...) error`
    - [x] **Create `binding/binding_test.go`**:
        - [x] Write comprehensive unit tests for all binding functions.
    - [x] **Create `binding/README.md`**:
        - [x] Provide a simple but complete usage example.
    - [x] **Add Form Value Binding**: Extend the package to support `application/x-w-form-urlencoded` and `multipart/form-data`. ([sketch/plan-binding-form-values.md](./sketch/plan-binding-form-values.md))
- **Test Logger**: Add a t.Logf-based logger to rakudatest. ([sketch/plan-test-logger.md](./sketch/plan-test-logger.md))
    - [x] Implement an `slog.Handler` that writes to `*testing.T` via `t.Logf`.
    - [x] Update `rakudatest.Do` to inject the test logger into the request context.
    - [x] Add tests to verify the logger is injected and works correctly.
    - [x] Refactor codebase to use `*slog.Logger` directly, removing custom interface.
    - [x] Centralize context helper functions in `context.go`.
- **Structured Error Responses for Binding**: Implemented `binding.Join` to provide detailed, structured JSON error responses for validation failures. ([sketch/plan-binding-join.md](./sketch/plan-binding-join.md))
    - [x] Define `binding.Error` and `binding.ValidationErrors` structs.
    - [x] Implement the `binding.Join` function to collect and wrap errors.
    - [x] Update `One`, `OnePtr`, `Slice`, and `SlicePtr` to return `*binding.Error`.
    - [x] Modify `responder.Error` to handle `*binding.ValidationErrors` and produce a detailed JSON response.
    - [x] Update `examples/simple-rest-api` to use `binding.Join`.
    - [x] Add `binding/binding_join_test.go` with tests for the new structured error responses.
- **Centralized Logging**: Implemented a centralized logging strategy. ([sketch/plan-sharing-logger.md](./sketch/plan-sharing-logger.md))
    - [x] **Implement Functional Options for Builder**:
        - [x] Create `BuilderConfig` struct and `BuilderOption` type.
        - [x] Implement `WithLogger(l *slog.Logger) BuilderOption`.
        - [x] Update `NewBuilder(...)` to accept `...BuilderOption`.
    - [x] **Implement Logger Injection Middleware**:
        - [x] In `Builder.Build()`, create a middleware that injects a logger into the request context.
        - [x] The middleware should add contextual attributes (`method`, `path`) to the logger.
        - [x] The middleware must check for a pre-existing logger (from `rakudatest`) and not overwrite it.
    - [x] **Refactor Components**:
        - [x] Remove the `Logger()` method from the `Responder` struct.
        - [x] Update `Responder` methods (`Error`, `JSON`, etc.) to get the logger directly from the context.
        - [x] Update `rakudamiddleware.Recovery` to get the logger from the context.
    - [x] **Implement Fallback with Warning**:
        - [x] Modify `LoggerFromContext` to fall back to `slog.Default()`.
        - [x] Use `sync.Once` to log a warning message the first time the fallback occurs.

## To Be Implemented

### Form Value Binding ([sketch/plan-binding-form-values.md](./sketch/plan-binding-form-values.md))
- [ ] Introduce a `Form` Source.
- [ ] Update value retrieval logic to use `r.PostForm`.
- [ ] No public API changes are needed.
- [ ] Write comprehensive tests for form binding.

### Structured Error Responses ([sketch/plan-binding-join.md](./sketch/plan-binding-join.md))
- [ ] Define `binding.Error` and `binding.ValidationErrors` structs.
- [ ] Implement the `binding.Join` function.
- [ ] Update core binding functions to return `*binding.Error`.
- [ ] Modify `responder.Error` to handle `*binding.ValidationErrors`.
- [ ] Update examples to use `binding.Join`.
- [ ] Add tests for new structured error responses.

### Binding Package ([sketch/plan-binding.md](./sketch/plan-binding.md))
- [ ] Create `binding/binding.go` with core types.
- [ ] Implement binding functions: `One`, `OnePtr`, `Slice`, `SlicePtr`.
- [ ] Create `binding/binding_test.go` with comprehensive tests.
- [ ] Create `binding/README.md` with a usage example.

### Centralized Logging ([sketch/plan-sharing-logger.md](./sketch/plan-sharing-logger.md))
- [ ] Implement functional options for `Builder` with `BuilderConfig` and `WithLogger`.
- [ ] Implement a logger injection middleware in `Builder.Build()`.
- [ ] Refactor `Responder` and `rakudamiddleware.Recovery` to get the logger from the context.
- [ ] Implement fallback to `slog.Default()` with a `sync.Once` warning.

### SSE Responder ([sketch/plan-sse-responder.md](./sketch/plan-sse-responder.md))
- [ ] Implement the `SSE` function in `responder.go`.
- [ ] Add `*Event` helper for sending named events.
- [ ] Add tests for SSE functionality in `responder_test.go`.
