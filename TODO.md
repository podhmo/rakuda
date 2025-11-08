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

## To Be Implemented

### Core Router Implementation (TDD)

The key design principle is the **two-stage separation**: configuration stage (Builder) and execution stage (http.Handler). Routes and middlewares can be defined in any order without affecting behavior.

- [x] **Builder Type with Tests**: Implement the `rakuda.Builder` type with internal configuration tree
  - [x] Test: Builder creation and basic structure
  - [x] Create `node` struct for configuration tree
  - [x] Implement `NewBuilder()` constructor
  - [x] Test: Route registration in any order produces consistent results
  - [x] Add basic route registration methods (`Get`, `Post`, `Put`, `Delete`, `Patch`)
- [-] **Build Process with Tests**: Implement the `Build()` method
  - [x] Test: Build creates immutable http.Handler
  - [x] Test: Order-independent route registration
  - [ ] DFS traversal of configuration tree
  - [ ] Context accumulation (path prefix and middleware chain)
  - [ ] Handler assembly with middleware wrapping
  - [x] Integration with `http.ServeMux`
- [ ] **Middleware Support with Tests**: Implement middleware functionality
  - [ ] Test: Global middleware application
  - [ ] Test: Scoped middleware application
  - [ ] Test: Middleware chain composition
  - [ ] `Use()` method for registering middlewares
  - [ ] Global middleware application
  - [ ] Scoped middleware application
  - [ ] Middleware chain composition
- [ ] **Route Grouping with Tests**: Implement route grouping functionality
  - [ ] Test: Nested route groups with path concatenation
  - [ ] Test: Middleware inheritance in nested groups
  - [ ] `Route(pattern, fn)` method for nested routes
  - [ ] `Group(fn)` method for middleware-only groups
  - [ ] Proper path concatenation
  - [ ] Middleware inheritance

### Example Applications
- [ ] **Simple REST API example**: Demonstrate basic usage
- [ ] **Middleware demonstration**: Show global and scoped middleware
- [ ] **Nested groups example**: Show route grouping patterns
