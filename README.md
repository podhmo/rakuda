# rakuda 🐪

`rakuda` is an HTTP router for Go, designed around a core philosophy of **compile-time safety** and **predictable lifecycle management**. It enforces a strict separation between route configuration and request handling through a dedicated `Builder` pattern, eliminating an entire class of runtime errors.

The name "rakuda" has two meanings:
- 「楽だ」(rakuda) - "Effortless" or "comfortable" in Japanese
- 「ラクダ」(rakuda) - "Camel" in Japanese 🐪

🚧 This library is currently under development.
- https://pkg.go.dev/github/podhmo/rakuda

## Features

- **Predictable Lifecycle**: Configuration of routes is completely separate from serving traffic. The router's state is immutable once built.
- **Declarative, Order-Independent API**: Declare routes and middlewares in any order without affecting final behavior.
- **Standard Library First**: Leverages `net/http` package, including path parameter support introduced in Go 1.22, for maximum compatibility and minimal dependencies.
- **Tree-Based Configuration**: Internal tree structure naturally maps to hierarchical RESTful API routes.
- **JSON Response Helper**: Built-in `Responder` for easy JSON responses with status code management.
- **Built-in Middlewares**: Recovery middleware for panic handling and CORS middleware for cross-origin requests.
- **Context-Aware Logging**: Logger and status code can be stored in request context for consistent error handling.
- **Debugging Tools**: `PrintRoutes` utility for visualizing all registered routes.

## Quick Start

The primary entry point is `rakuda.Builder`, which is used to configure routes and then build an immutable `http.Handler`.

```go
package main

import (
	"net/http"

	"github.com/podhmo/rakuda"
	"github.com/podhmo/rakuda/rakudamiddleware"
)

func main() {
    // Create a new builder for route configuration
    b := rakuda.NewBuilder()
    responder := rakuda.NewResponder()
    
    // Add global middleware
    b.Use(rakudamiddleware.Recovery)
    
    // Define routes
    b.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        responder.JSON(w, r, map[string]string{
            "message": "Welcome to rakuda!",
        })
    }))
    
    b.Get("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.PathValue("id")
        responder.JSON(w, r, map[string]string{
            "id": userID,
            "name": "John Doe",
        })
    }))
    
    // Build the immutable handler
    handler, err := b.Build()
    if err != nil {
        panic(err)
    }
    
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

`rakuda` uses Go 1.22's native path parameter support. Parameters are retrieved directly from the request:

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

handler, err := b.Build()
if err != nil {
    panic(err)
}
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

### JSON Responses with Responder

`rakuda` provides a `Responder` type for easy JSON response handling with built-in error logging:

```go
responder := rakuda.NewResponder()

b.Get("/users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("id")
    
    // Set status code in context
    r = rakuda.WithStatusCode(r, http.StatusOK)
    
    // Respond with JSON
    responder.JSON(w, r, map[string]string{
        "id": userID,
        "name": "John Doe",
    })
}))
```

The `Responder` automatically:
- Sets the correct `Content-Type` header
- Encodes data to JSON
- Logs encoding errors using the logger from context (or a default logger)
- Respects the status code set in the request context

### Simplified Handlers with `Lift`

For handlers that simply return data and an error, `rakuda` provides a `Lift` function. This generic function converts a handler of the form `func(*http.Request) (T, error)` into a standard `http.Handler`, automating JSON encoding and error handling.

- **On success**: The returned data is automatically encoded as a JSON response with a `200 OK` status.
- **On error**:
    - If the error provides a `StatusCode() int` method (like `rakuda.APIError`), that status code is used.
    - Otherwise, a generic `500 Internal Server Error` is returned to the client, while the original error is logged internally.

```go
// Define a response struct
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// This function matches the signature required by Lift
func GetUser(r *http.Request) (User, error) {
    userID := r.PathValue("id")
    if userID == "" {
        // Return an error with a specific status code
        return User{}, rakuda.NewAPIError(http.StatusBadRequest, errors.New("user ID is required"))
    }

    // On success, just return the data
    return User{ID: userID, Name: "John Doe"}, nil
}

// Use Lift to create the http.Handler
b.Get("/users/{id}", rakuda.Lift(responder, GetUser))
```

This pattern simplifies handler logic by removing the boilerplate of response writing and error checking.

### Built-in Middlewares

#### Recovery Middleware

The `Recovery` middleware catches panics, logs them with stack traces, and returns a 500 error:

```go
b := rakuda.NewBuilder()

// Apply recovery globally
b.Use(rakudamiddleware.Recovery)

b.Get("/panic", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    panic("something went wrong")  // Will be caught and logged
}))
```

#### CORS Middleware

The `CORS` middleware handles Cross-Origin Resource Sharing with configurable options:

```go
b := rakuda.NewBuilder()

// Use default permissive CORS settings
b.Use(rakudamiddleware.CORS(nil))

// Or configure CORS explicitly
b.Use(rakudamiddleware.CORS(&rakudamiddleware.CORSConfig{
    AllowedOrigins: []string{"https://example.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge: 3600,
}))
```

### Context Helpers

Store logger and status code in request context for consistent handling:

```go
// Store a logger in context (typically done in middleware)
r = rakuda.WithLogger(r, logger)

// Set status code (can be done anywhere in the handler chain)
r = rakuda.WithStatusCode(r, http.StatusCreated)
```

### Custom 404 Handler

Set a custom handler for routes that don't match:

```go
b := rakuda.NewBuilder()

b.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte("Page not found"))
}))
```

If not set, a default JSON 404 response is used.

### Debugging: Print Routes

Use `PrintRoutes` to display all registered routes:

```go
rakuda.PrintRoutes(os.Stdout, builder)
// Output:
// GET   /
// GET   /users/{id}
// POST  /users
```

This is useful for debugging and documentation. Many example applications include a `-proutes` flag to display routes without starting the server.

## Design Philosophy

For detailed information about the design decisions and architecture, see [docs/router-design.md](./docs/router-design.md).

Key design principles:

- **Fail Fast**: Configuration errors are caught at compile-time or application startup, not at runtime
- **Immutability**: Once built, the router cannot be modified
- **No Magic**: Clear, explicit API with predictable behavior
- **Standard Compliance**: Full compatibility with `net/http` ecosystem

## Examples

The repository includes several example applications in the `examples/` directory:

- **[simple-rest-api](./examples/simple-rest-api)**: Basic REST API with path parameters
- **[middleware-demonstration](./examples/middleware-demonstration)**: Shows Recovery middleware and panic handling
- **[spa-with-embed](./examples/spa-with-embed)**: Single Page Application with embedded static files

Each example can be run with `go run` and many include a `-proutes` flag to display registered routes.

## Requirements

- Go 1.24 or later 

## Installation

```bash
go get github.com/podhmo/rakuda
```

## License

MIT License - see [LICENSE](./LICENSE) file for details.
