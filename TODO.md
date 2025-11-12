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
    - [x] **Add Form Value Binding**: Extend the package to support `application/x-www-form-urlencoded` and `multipart/form-data`. ([sketch/plan-binding-form-values.md](./sketch/plan-binding-form-values.md))
- **Test Logger**: Add a t.Logf-based logger to rakudatest. ([sketch/plan-test-logger.md](./sketch/plan-test-logger.md))
    - [x] Implement an `slog.Handler` that writes to `*testing.T` via `t.Logf`.
    - [x] Update `rakudatest.Do` to inject the test logger into the request context.
    - [x] Add tests to verify the logger is injected and works correctly.
    - [x] Refactor codebase to use `*slog.Logger` directly, removing custom interface.
    - [x] Centralize context helper functions in `context.go`.

## To Be Implemented

### Core Router Implementation (TDD)

The key design principle is the **two-stage separation**: configuration stage (Builder) and execution stage (http.Handler). Routes and middlewares can be defined in any order without affecting behavior.

- [x] **Builder Type with Tests**: Implement the `rakuda.Builder` type with internal configuration tree
  - [x] Test: Builder creation and basic structure
  - [x] Create `node` struct for configuration tree
  - [x] Implement `NewBuilder()` constructor
  - [x] Test: Route registration in any order produces consistent results
  - [x] Add basic route registration methods (`Get`, `Post`, `Put`, `Delete`, `Patch`)
- [x] **Build Process with Tests**: Implement the `Build()` method
  - [x] Test: Build creates immutable http.Handler
  - [x] Test: Order-independent route registration
  - [x] DFS traversal of configuration tree
  - [x] Context accumulation (path prefix and middleware chain)
  - [x] Handler assembly with middleware wrapping
  - [x] Integration with `http.ServeMux`
- [x] **Middleware Support with Tests**: Implement middleware functionality
  - [x] Test: Global middleware application
  - [x] Test: Scoped middleware application
  - [x] Test: Middleware chain composition
  - [x] `Use()` method for registering middlewares
  - [x] Global middleware application
  - [x] Scoped middleware application
  - [x] Middleware chain composition
- [-] **Route Grouping with Tests**: Implement route grouping functionality
  - [x] Test: Nested route groups with path concatenation
  - [x] Test: Middleware inheritance in nested groups
  - [x] `Route(pattern, fn)` method for nested routes
  - [x] `Group(fn)` method for middleware-only groups
  - [x] Proper path concatenation
  - [x] Middleware inheritance
- [x] **Not Found Handler**: Support for custom 404 handlers.
  - [x] `NotFound(handler)` method on Builder.
  - [x] Default JSON 404 response if no handler is provided.
  - [x] Test: Default handler behavior.
  - [x] Test: Custom handler behavior.
- [x] **Route Conflict Handling**: Add configurable behavior for duplicate route registrations.
  - [x] `OnConflict` field in `Builder` (`Warn` or `Error`).
  - [x] `Build()` method returns an error on conflict when configured to do so.
  - [x] Add tests for both `Warn` and `Error` behaviors.

### SSE Responder
- [x] **Server-Sent Events (SSE) Support**: Add a responder for streaming data.
  - [x] `SSE` function to handle `text/event-stream` responses.
  - [x] `Event` struct for sending named events.
  - [x] Add tests for SSE functionality.

### Example Applications
- [x] **Simple REST API example**: Demonstrate basic usage
- [x] **Middleware demonstration**: Show global and scoped middleware
- [ ] **Nested groups example**: Show route grouping patterns

### CLI
- [x] **proutes utility**: Add a utility to display registered handlers.
    - [x] Add `-proutes` flag to the example application.

### Structured Error Responses for Binding
- [ ] **Implement `binding.Join`**: Create a `Join` function in the `binding` package to aggregate validation errors into a structured format. ([sketch/plan-binding-join.md](./sketch/plan-binding-join.md))
    - [ ] Define `binding.Error` and `binding.ValidationErrors` structs.
    - [ ] Implement the `binding.Join` function to collect and wrap errors.
    - [ ] Update `One`, `OnePtr`, `Slice`, and `SlicePtr` to return `*binding.Error`.
    - [ ] Modify `responder.Error` to handle `*binding.ValidationErrors` and produce a detailed JSON response.
    - [ ] Update `examples/simple-rest-api` to use `binding.Join`.
    - [ ] Add `binding/binding_join_test.go` with tests for the new structured error responses.
