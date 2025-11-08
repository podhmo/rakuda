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

### Example Applications
- [x] **Simple REST API example**: Demonstrate basic usage
- [x] **Middleware demonstration**: Show global and scoped middleware
- [ ] **Nested groups example**: Show route grouping patterns
