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

### Core Router Implementation
- [ ] **Builder Type**: Implement the `rakuda.Builder` type with internal configuration tree
  - [ ] Create `node` struct for configuration tree
  - [ ] Implement `NewBuilder()` constructor
  - [ ] Add basic route registration methods (`Get`, `Post`, `Put`, `Delete`, `Patch`)
- [ ] **Build Process**: Implement the `Build()` method
  - [ ] DFS traversal of configuration tree
  - [ ] Context accumulation (path prefix and middleware chain)
  - [ ] Handler assembly with middleware wrapping
  - [ ] Integration with `http.ServeMux`
- [ ] **Middleware Support**: Implement middleware functionality
  - [ ] `Use()` method for registering middlewares
  - [ ] Global middleware application
  - [ ] Scoped middleware application
  - [ ] Middleware chain composition
- [ ] **Route Grouping**: Implement route grouping functionality
  - [ ] `Route(pattern, fn)` method for nested routes
  - [ ] `Group(fn)` method for middleware-only groups
  - [ ] Proper path concatenation
  - [ ] Middleware inheritance

### Testing Infrastructure
- [ ] **Unit Tests**: Create comprehensive unit tests
  - [ ] Builder configuration tests
  - [ ] Middleware chain tests
  - [ ] Route grouping tests
  - [ ] Path parameter tests
- [ ] **Integration Tests**: Create integration tests
  - [ ] Full request/response cycle tests
  - [ ] Multiple route group tests
  - [ ] Middleware ordering tests
- [ ] **Example Applications**: Create example applications
  - [ ] Simple REST API example
  - [ ] Middleware demonstration
  - [ ] Nested groups example

### Documentation and Tooling
- [ ] **API Documentation**: Generate and publish Go package documentation
- [ ] **Tutorial**: Create step-by-step tutorial for common use cases
- [ ] **Performance Benchmarks**: Benchmark against other popular routers
- [ ] **CI/CD**: Set up automated testing and linting

### Advanced Features (Future)
- [ ] **Error Handling**: Implement error handling utilities
- [ ] **Request Context Helpers**: Add utilities for common request context patterns
- [ ] **Testing Helpers**: Create test utilities for testing handlers
- [ ] **Metrics Integration**: Add optional metrics hooks
