# Design Doc: Centralized Logging in Rakuda

## Background

Currently, logging in the Rakuda framework is decentralized and inconsistent. Key components instantiate their own loggers with default configurations if a logger is not found in the `http.Request` context.

- **`Responder`**: `NewResponder()` creates a default `slog.Logger` that writes JSON to `os.Stderr`. The `Responder.Logger()` method falls back to this instance.
- **`rakudamiddleware.Recovery`**: The recovery middleware also creates its own default `slog.Logger` in the event of a panic if no logger is available in the context.

This approach leads to several issues:
- **Inconsistent Configuration**: There is no single point to configure the application's logger. Setting a global log level, changing the output format (e.g., from JSON to text), or directing logs to a file is not possible.
- **Disconnected Logging**: Logs from different parts of the framework may have different formats or destinations, making debugging difficult.
- **Testing Challenges**: While `rakudatest` can inject a logger for handler-level tests, testing the logging behavior of middleware itself is cumbersome.

## Problem Statement

The lack of a unified logging strategy prevents users from effectively managing and controlling log output across their entire application. A centralized mechanism is needed to configure a logger once and have it be respected by all framework components.

## Goals

1.  **Centralize Logger Configuration**: Provide a single point for users to configure the logger for the entire application.
2.  **Consistent Propagation**: Ensure the configured logger is consistently available in the request context for all handlers and middlewares.
3.  **Improve Observability**: Automatically enrich logs with contextual request information (e.g., method, path).
4.  **Maintain Testability**: The solution must not break the ability of `rakudatest` to inject custom loggers for testing purposes.

## Proposed Design

The proposed solution centers around making the `Builder` the owner of the logging configuration. A new middleware, implicitly added during the build process, will be responsible for injecting the configured logger into the request context.

### 1. Builder Configuration

The `Builder` will be enhanced with new options to control logging:

```go
// In builder.go

type Builder struct {
    // ... existing fields
    logger         *slog.Logger
    logLevel       slog.Leveler // For creating a default logger if one isn't provided
}

func NewBuilder() *Builder {
    return &Builder{
        // ...
        logLevel: slog.LevelInfo, // Default log level
    }
}

// SetLogger allows the user to provide a pre-configured logger instance.
func (b *Builder) SetLogger(l *slog.Logger) {
    b.logger = l
}

// SetLogLevel sets the minimum level for the default logger.
// This is ignored if a custom logger is provided via SetLogger.
func (b *Builder) SetLogLevel(level slog.Leveler) {
    b.logLevel = level
}
```

### 2. Logger Injection Middleware

During `Builder.Build()`, a special logger middleware will be constructed and prepended to the root-level middleware chain. This ensures it runs for every single request before any other middleware or handler.

```go
// In builder.go's Build() method (conceptual)

func (b *Builder) Build() (http.Handler, error) {
    // ...

    // 1. Determine the logger to use
    logger := b.logger
    if logger == nil {
        logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: b.logLevel}))
    }

    // 2. Create the injection middleware
    loggingMiddleware := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Add contextual attributes
            requestLogger := logger.With(
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
            )
            // Inject into context
            ctx := NewContextWithLogger(r.Context(), requestLogger)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }

    // 3. Prepend this middleware to the global chain before building the router
    // ... build logic continues ...
}
```

### 3. Refactoring Core Components

With the guarantee that a logger will always be in the context, we can simplify other components.

**`Responder`**:
The `Responder` will no longer create or store a default logger. It will exclusively rely on the context.

```go
// In responder.go

type Responder struct {
    // No logger field
}

func NewResponder() *Responder {
    return &Responder{}
}

// Logger strictly retrieves the logger from the context.
// If it's missing, it falls back to the global default slog logger,
// but the design assumes it will always be present.
func (r *Responder) Logger(ctx context.Context) *slog.Logger {
    if logger, ok := LoggerFromContext(ctx); ok {
        return logger
    }
    return slog.Default() // A safe fallback, but indicates a configuration issue.
}
```

**`rakudamiddleware.Recovery`**:
The `Recovery` middleware will be simplified to remove its custom logger creation.

```go
// In rakudamiddleware/recover.go

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Logger is now guaranteed to be in the context
                logger, _ := rakuda.LoggerFromContext(r.Context())
                logger.ErrorContext(r.Context(), "panic recovered", "error", err, "stack", string(debug.Stack()))

                responder := rakuda.NewResponder()
                responder.Error(w, r, http.StatusInternalServerError, http.ErrAbortHandler)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### 4. Impact on `rakudatest`

This design does not negatively impact `rakudatest`. Test helpers will continue to work as expected. When a user creates a test request (`httptest.NewRequest`), they can inject a test-specific logger (e.g., one that writes to a buffer) into the request context before passing it to the handler. The handler will find and use the test logger, overriding the one configured in the `Builder`.

## Open Questions

1.  **Fallback Behavior**: What should happen if `LoggerFromContext` fails? The current proposal is to fall back to `slog.Default()`. Should it panic instead to enforce correct configuration? A fallback is safer, but a panic makes misconfiguration more obvious during development.
2.  **Contextual Attributes**: Is `method` and `path` sufficient? Should other attributes like `remote_addr` or `request_id` (from a hypothetical middleware) be included? For now, method and path are a good starting point.
