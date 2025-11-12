# Design Doc: Centralized Logging in Rakuda (Revised)

## Background

Currently, logging in the Rakuda framework is decentralized. Key components like `Responder` and `rakudamiddleware.Recovery` instantiate their own `slog.Logger` with default configurations if a logger is not found in the `http.Request` context. This leads to inconsistent configuration, disconnected logging, and challenges in testing.

## Problem Statement

The lack of a unified logging strategy prevents users from effectively managing and controlling log output. A centralized mechanism is needed to configure a logger once and have it be respected by all framework components.

## Goals

1.  **Centralize Logger Configuration**: Provide a single, flexible point for users to configure the application logger.
2.  **Consistent Propagation**: Ensure the configured logger is consistently available in the request context for all handlers and middlewares.
3.  **Improve Observability**: Automatically enrich logs with contextual request information (e.g., method, path).
4.  **Maintain Testability**: The solution must not break the ability of `rakudatest` to inject custom loggers for testing purposes.

## Proposed Design

The revised design makes the `Builder` the owner of the logging configuration, using the **functional options pattern** for a clean and extensible API. A new middleware, implicitly added during the build process, will be responsible for injecting the configured logger into the request context.

### 1. Builder Configuration (Functional Options)

The `Builder` will hold an unexported `config` struct. It will be configured at creation time via optional functions (`func(*BuilderConfig)`), preventing its internal state from being modified after instantiation.

```go
// In builder.go

// BuilderConfig holds the configuration for a Builder.
type BuilderConfig struct {
	Logger *slog.Logger
}

// BuilderOption is a function that configures a Builder.
type BuilderOption func(*BuilderConfig)

// WithLogger is a BuilderOption to provide a pre-configured logger.
// Users can create a logger with a slog.LevelVar to control the level dynamically.
func WithLogger(l *slog.Logger) BuilderOption {
	return func(c *BuilderConfig) {
		c.Logger = l
	}
}

// Builder struct
type Builder struct {
	config BuilderConfig
	// ... other existing fields
}

// NewBuilder creates a new Builder with the given options.
func NewBuilder(options ...BuilderOption) *Builder {
	var config BuilderConfig
	for _, opt := range options {
		opt(&config)
	}

	b := &Builder{
		config: config,
		// ...
	}
	return b
}
```

This approach eliminates the need for `SetLogger` or `SetLogLevel` methods. Users are responsible for creating their logger, including its `slog.LevelVar` if dynamic control is desired.

### 2. Logger Injection Middleware

During `Builder.Build()`, a special logger middleware will be constructed and prepended to the root-level middleware chain.

```go
// In builder.go's Build() method (conceptual)

func (b *Builder) Build() (http.Handler, error) {
    // ...

    // 1. Determine the logger to use
    logger := b.config.Logger
    if logger == nil {
        // Create a default logger if none was provided.
        // The default level is Info.
        logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
    }

    // 2. Create the injection middleware
    loggingMiddleware := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestLogger := logger.With(
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
            )
            ctx := NewContextWithLogger(r.Context(), requestLogger)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }

    // 3. Prepend this middleware to the global chain before building the router
    // ... build logic continues ...
}
```

### 3. Refactoring Core Components

With the guarantee that a logger will almost always be in the context, we can simplify other components and establish a clear fallback mechanism.

#### Fallback Behavior

If `LoggerFromContext` fails to find a logger, components will fall back to `slog.Default()`. To alert developers of a potential misconfiguration, a warning should be logged **once** per application lifecycle.

```go
// in context.go

var logFallbackOnce sync.Once

func LoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
    logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
    if !ok {
        logFallbackOnce.Do(func() {
            slog.Default().Warn("Logger not found in context; falling back to slog.Default(). This may indicate a misconfiguration or a handler used outside of the Rakuda router.")
        })
        return slog.Default(), false
    }
    return logger, true
}
```

**`Responder`**:
The `Responder` will be simplified, removing the `Logger` method entirely and relying on the behavior defined above.

```go
// In responder.go

type Responder struct {
    // No logger field
}

func NewResponder() *Responder {
    return &Responder{}
}

// Methods like Error() and JSON() will now call LoggerFromContext directly.
func (r *Responder) Error(w http.ResponseWriter, req *http.Request, statusCode int, err error) {
    ctx := req.Context()
    logger, _ := LoggerFromContext(ctx) // Directly get logger, fallback is handled inside.
    // ... logging and response logic ...
}
```

**`rakudamiddleware.Recovery`**:
The `Recovery` middleware will also be simplified.

```go
// In rakudamiddleware/recover.go

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // Logger is now retrieved from context, with a safe fallback.
                logger, _ := rakuda.LoggerFromContext(r.Context())
                logger.ErrorContext(r.Context(), "panic recovered", "error", err, "stack", string(debug.Stack()))

                // ...
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### 4. Impact on `rakudatest` and Further Considerations

This design remains fully compatible with `rakudatest`. Tests can inject a custom logger into the request context, and that logger will be found and used by `LoggerFromContext`, correctly overriding any application-level logger.

However, a key implementation detail must be handled carefully:
- **Test Logger Precedence**: The logger injection middleware, which is prepended by the `Builder`, runs on every request. There is a risk that this middleware could overwrite (or shadow) a custom logger that was injected into the context by a test helper like `rakudatest.Do`. The middleware's implementation must first check if a logger already exists in the context. If one is present, it should *not* be replaced, ensuring that the test-specific logger always takes precedence.

## Open Questions

1.  **Contextual Attributes**: Is `method` and `path` sufficient? For now, this is a good, non-breaking starting point. More advanced logging (e.g., with request IDs) can be added by users with custom middleware.
