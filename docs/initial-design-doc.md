

# Design Document: `rakuda` HTTP Router

- **Author:** Gemini
- **Status:** Proposed
- **Date:** 2025-09-07

## 1. Abstract

This document outlines the internal design and architecture of `rakuda`, a new HTTP router for Go. The primary goal of `rakuda` is to provide a **type-safe** and **predictable** routing experience by enforcing a strict separation between the configuration phase and the execution phase. This is achieved through a dedicated `Builder` type for route definition and the final output of a standard `http.Handler`. This design eliminates a class of common runtime errors, such as using a router before its configuration is complete.

## 2. Goals and Motivation

The Go ecosystem has many routers, but many either allow for runtime errors due to ambiguous lifecycles or have complex APIs. The motivation for `rakuda` is to address these issues with the following core goals:

- **Type Safety:** The router's lifecycle must be enforced by the Go compiler. It should be impossible to compile code that uses a router before it has been explicitly built.
- **Predictable Lifecycle:** The configuration of routes (the "what") should be completely separate from the serving of traffic (the "how"). The router's state should be immutable once built.
- **Declarative, Order-Independent API:** Developers should be able to declare routes and middlewares in any order without affecting the final behavior.
- **Leverage Standard Library:** Maximize the use of the standard `net/http` package, including the path parameter support introduced in Go 1.22, to ensure compatibility and minimize dependencies.
- **Flexible Middleware Scoping:** Support applying middlewares globally, to specific route groups, or to nested groups.

## 3. Proposed Architecture

### 3.1. Core Components: `Builder` and `http.Handler`

The entire design hinges on the strict separation of two types:

- **`rakuda.Builder`**: This is the sole object responsible for configuration. It exposes methods like `Get`, `Post`, `Use`, `Route`, etc. Critically, it **does not** implement the `http.Handler` interface. Its only responsibility is to build and maintain an internal configuration tree.

- **`http.Handler`**: This is the standard library interface. The `Builder.Build()` method is the factory for this type. The returned handler is an immutable, ready-to-use component that encapsulates all the defined routing logic.

This separation forces the developer into a safe, two-step process: first, configure the `Builder`; second, `Build()` it into an `http.Handler` to be used by the server. An attempt to pass the `Builder` to `http.ListenAndServe` will result in a compile-time error.

### 3.2. Internal Data Structure: The Configuration Tree

To support flexible middleware scoping and nested routes (`/api/v1/...`), the `Builder` does not store routes in a flat list. Instead, it maintains a **tree of configuration nodes**.

- Each `Builder` instance holds a reference to a `node` in this tree.
- `NewBuilder()` creates the `root` node.
- `Route(pattern, fn)` and `Group(fn)` create a new `child` node under the current node. The function `fn` receives a new `Builder` instance that points to this new child node.

A simplified representation of a `node` struct would be:

```go
type node struct {
    pattern     string
    middlewares []func(http.Handler) http.Handler
    handlers    []handlerRegistration
    children    []*node
}
```

This tree structure naturally maps to the hierarchical nature of RESTful API routes.

### 3.3. The Build Process

All the "magic" happens within the `Build()` method. The method calls on the `Builder` (`Get`, `Use`, etc.) are lightweight operations that simply populate the configuration tree. `Build()` orchestrates the following process:

1.  **Traversal**: It performs a Depth-First Search (DFS) traversal of the entire configuration tree, starting from the root.
2.  **Context Accumulation**: As it traverses, it keeps track of the accumulated path prefix (e.g., `/api` + `/v1`) and the inherited middleware chain.
3.  **Handler Assembly**: For each node it visits, it:
    a. Creates the full middleware chain for that node by appending the node's own middlewares to the chain inherited from its parent.
    b. Iterates through all handlers registered on that node.
    c. For each handler, it applies the full middleware chain.
    d. It registers the final, wrapped handler with a new `http.ServeMux` instance, using the full, concatenated path pattern (e.g., `GET /api/v1/users/{id}`).
4.  **Finalization**: Once the traversal is complete, the populated `http.ServeMux` (which is itself an `http.Handler`) is returned. The `Builder` is then marked as "built" to prevent further modification.

This deferred assembly process is what enables the order-independent API. The order of `Use()` and `Get()` calls only matters relative to each other *after* the build process has categorized and applied them.

### 3.4. Path Parameters

`rakuda` will **not** implement its own path parameter parsing logic. It will rely entirely on the native capabilities of `net/http` introduced in Go 1.22. Handlers are expected to retrieve path parameters directly from the request object.

**Example:**

```go
// Registration
b.Get("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Retrieval inside the handler
    userID := r.PathValue("id")
    // ...
}))
```

## 4. Rejected Alternatives

Several alternative designs were considered and rejected. Understanding these trade-offs is crucial to understanding the proposed architecture.

### 4.1. Alternative 1: Lazy Initialization

- **Description**: The router would implement `http.Handler`, but the actual building would be deferred until the first time `ServeHTTP` is called, likely using a `sync.Once`.
- **Reason for Rejection**: This is a poor developer experience. Configuration errors (e.g., a typo in a route, an invalid pattern) would only surface at runtime when the first request hits the server, not at application startup. This violates the principle of failing fast.

### 4.2. Alternative 2: Combined Router/Builder Type

- **Description**: A single `Router` type would both implement `http.Handler` and expose configuration methods. A `Build()` method might exist, but its use wouldn't be enforced by the compiler.
- **Reason for Rejection**: This is the most common pattern, but it's prone to error. A developer could easily forget to call `Build()` and pass the unconfigured or partially configured router to the server. This would lead to runtime panics (if `ServeHTTP` checks for a built state) or silent failures (404s for all routes). The proposed design solves this at compile time.

### 4.3. Alternative 3: Flat Slice and Sorting

- **Description**: Instead of a tree, all registrations (middlewares and handlers) would be appended to a single flat slice. Each entry would have a "type" or "priority" (e.g., middleware=1, handler=2). During `Build()`, this slice would be sorted to ensure middlewares are processed before handlers.
- **Reason for Rejection**: This approach is too simplistic. It cannot handle scoped middlewares. For example, you cannot apply a middleware *only* to routes under `/api`. All middlewares would effectively become global, severely limiting the router's utility for any non-trivial application. The tree structure is essential for this functionality.
