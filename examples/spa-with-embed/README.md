# SPA with Embed Example

This example demonstrates how to build a Single Page Application (SPA) with rakuda, featuring:

- **go:embed** for serving static files
- **CORS middleware** for cross-origin requests
- **Multiple route groups** with scoped middleware
- **Authentication and authorization** middleware
- **Path parameters** for dynamic routing
- **JSON API responses** using the Responder

## Features Demonstrated

### 1. go:embed Integration

Static files (HTML, CSS, JavaScript) are embedded directly into the binary using Go's `embed` package:

```go
//go:embed static/*
var staticFiles embed.FS
```

This means the application is a single binary with no external dependencies for static assets.

### 2. CORS Middleware

The example includes a comprehensive CORS middleware that handles:
- Preflight requests (OPTIONS)
- Multiple allowed origins
- Configurable methods and headers
- Credentials support
- Max-Age caching

```go
api.Use(rakuda.CORS(&rakuda.CORSConfig{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
    AllowCredentials: false,
    MaxAge:           3600,
}))
```

### 3. Route Grouping with Scoped Middleware

The example demonstrates rakuda's powerful route grouping:

```go
builder.Route("/api", func(api *rakuda.Builder) {
    api.Use(rakuda.CORS(config))
    api.Use(loggingMiddleware())
    
    // Public routes
    api.Route("/public", func(public *rakuda.Builder) {
        public.Get("/info", handler)
    })
    
    // Protected user routes
    api.Route("/users", func(users *rakuda.Builder) {
        users.Use(authMiddleware())
        users.Get("/current", handler)
        users.Get("/{id}", handler)
    })
    
    // Admin-only routes
    api.Route("/admin", func(admin *rakuda.Builder) {
        admin.Use(authMiddleware())
        admin.Use(adminOnlyMiddleware())
        admin.Get("/stats", handler)
    })
})
```

Middleware is applied hierarchically:
- **All API routes** get CORS and logging
- **User routes** additionally get authentication
- **Admin routes** get authentication + admin authorization

### 4. Custom Middleware

The example includes three custom middleware implementations:

- **loggingMiddleware**: Logs request details and duration
- **authMiddleware**: Extracts and validates authentication tokens
- **adminOnlyMiddleware**: Enforces admin-only access

### 5. Path Parameters

Dynamic routing with path parameters:

```go
users.Get("/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("id")
    // Use userID...
}))
```

## Running the Example

### Basic Usage

```bash
cd examples/spa-with-embed
go run main.go
```

Then open your browser to http://localhost:8080

### Custom Port

```bash
go run main.go -port 3000
```

### View Routes

```bash
go run main.go -proutes
```

This will display all registered routes:

```
GET    /
GET    /static/{path...}
GET    /api/public/info
GET    /api/users/current
GET    /api/users/{id}
POST   /api/users/
GET    /api/admin/stats
```

## API Endpoints

### Public Endpoints

- `GET /api/public/info` - Returns API information (no auth required)

### User Endpoints (requires Authorization header)

- `GET /api/users/current` - Get current authenticated user
- `GET /api/users/{id}` - Get user by ID
- `POST /api/users/` - Create a new user

### Admin Endpoints (requires admin token)

- `GET /api/admin/stats` - Get system statistics

## Testing with curl

### Public endpoint:
```bash
curl http://localhost:8080/api/public/info
```

### User endpoint with auth:
```bash
curl -H "Authorization: Bearer demo-token-12345" \
     http://localhost:8080/api/users/current
```

### Admin endpoint (will fail without admin token):
```bash
curl -H "Authorization: Bearer demo-token-12345" \
     http://localhost:8080/api/admin/stats
```

### Admin endpoint (with admin token):
```bash
curl -H "Authorization: Bearer admin-token-67890" \
     http://localhost:8080/api/admin/stats
```

### Get user by ID:
```bash
curl http://localhost:8080/api/users/123
```

## CORS Testing

The CORS middleware is configured to allow all origins (`*`). To test cross-origin requests:

1. Open the SPA in your browser at http://localhost:8080
2. Open the browser's developer console
3. The JavaScript code will make requests to the API endpoints
4. Check the Network tab to see the CORS headers in action

For preflight requests:
```bash
curl -X OPTIONS http://localhost:8080/api/users/current \
     -H "Origin: http://example.com" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Authorization" \
     -v
```

## Building

To build a single binary:

```bash
go build -o spa-example
./spa-example
```

All static files are embedded in the binary, so it can be deployed without any external dependencies.

## Key Takeaways

This example showcases rakuda's strengths:

1. **Type Safety**: The Builder pattern prevents using unconfigured routers
2. **Declarative API**: Routes and middleware can be declared in any order
3. **Middleware Composition**: Middleware is inherited and composed naturally
4. **Standard Library Integration**: Uses `net/http` and `embed` packages
5. **Clean Architecture**: Clear separation between configuration and execution

The example demonstrates a real-world pattern for building modern web applications with Go.
