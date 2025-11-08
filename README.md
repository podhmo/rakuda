# rakuda 🐪

`rakuda` is an HTTP router for Go, designed around a core philosophy of **compile-time safety** and **predictable lifecycle management**. It enforces a strict separation between route configuration and request handling through a dedicated `Builder` pattern, eliminating an entire class of runtime errors.

The name "rakuda" has two meanings:
- 「楽だ」(rakuda) - "Effortless" or "comfortable" in Japanese
- 「ラクダ」(rakuda) - "Camel" in Japanese 🐪

🚧 This library is currently under development.

## Features

- **Predictable Lifecycle**: Configuration of routes is completely separate from serving traffic. The router's state is immutable once built.
- **Declarative, Order-Independent API**: Declare routes and middlewares in any order without affecting final behavior.
- **Standard Library First**: Leverages `net/http` package, including path parameter support introduced in Go 1.24, for maximum compatibility and minimal dependencies.
- **Tree-Based Configuration**: Internal tree structure naturally maps to hierarchical RESTful API routes.

## Quick Start

The primary entry point is `rakuda.Builder`, which is used to configure routes and then build an immutable `http.Handler`.

```go
package main

import (
    "fmt"
    "net/http"
    
    "github.com/podhmo/rakuda"
)

func main() {
    // Create a new builder for route configuration
    b := rakuda.NewBuilder()
    
    // Define routes
    b.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Welcome to rakuda!")
    }))
    
    b.Get("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.PathValue("id")
        fmt.Fprintf(w, "User ID: %s\n", userID)
    }))
    
    // Build the immutable handler
    handler := b.Build()
    
    // Start the server
    http.ListenAndServe(":8080", handler)
}
```

## Core Concepts

### Builder Pattern

`rakuda` uses a two-phase approach:

1. **Configuration Phase**: Use `rakuda.Builder` to define routes and middlewares
2. **Execution Phase**: Call `Build()` to create an immutable `http.Handler`

This separation is enforced at compile-time. You cannot pass a `Builder` to `http.ListenAndServe` - it will fail to compile.

### Path Parameters

`rakuda` uses Go 1.24's native path parameter support. Parameters are retrieved directly from the request:

```go
b.Get("/posts/{postID}/comments/{commentID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    postID := r.PathValue("postID")
    commentID := r.PathValue("commentID")
    // Handle the request
}))
```

### Route Groups and Middleware

Apply middlewares to specific route groups using nested builders:

```go
b := rakuda.NewBuilder()

// Global middleware
b.Use(loggingMiddleware)

// API v1 group
b.Route("/api/v1", func(api *rakuda.Builder) {
    // Middleware scoped to /api/v1
    api.Use(authMiddleware)
    
    api.Get("/users", listUsersHandler)
    api.Post("/users", createUserHandler)
    
    // Nested group
    api.Route("/admin", func(admin *rakuda.Builder) {
        admin.Use(adminOnlyMiddleware)
        admin.Get("/stats", adminStatsHandler)
    })
})

handler := b.Build()
```

#### Order-Independent Configuration

One of `rakuda`'s key features is its **order-independent API**. You can declare routes and middlewares in any order within the same scope without affecting the final behavior:

```go
// These two configurations produce identical results:

// Configuration 1: middleware first
b.Route("/api", func(api *rakuda.Builder) {
    api.Use(authMiddleware)
    api.Get("/users", listUsersHandler)
})

// Configuration 2: route first
b.Route("/api", func(api *rakuda.Builder) {
    api.Get("/users", listUsersHandler)
    api.Use(authMiddleware)
})
```

This is possible because the actual middleware chain is assembled during the `Build()` phase, not at the time of declaration. The builder collects all configuration declaratively and processes it consistently, regardless of the order in which you register routes and middlewares.

## Design Philosophy

For detailed information about the design decisions and architecture, see [docs/router-design.md](./docs/router-design.md).

Key design principles:

- **Fail Fast**: Configuration errors are caught at compile-time or application startup, not at runtime
- **Immutability**: Once built, the router cannot be modified
- **No Magic**: Clear, explicit API with predictable behavior
- **Standard Compliance**: Full compatibility with `net/http` ecosystem

## Requirements

- Go 1.24 or later (for native path parameter support)

## Installation

```bash
go get github.com/podhmo/rakuda
```

## License

MIT License - see [LICENSE](./LICENSE) file for details.
