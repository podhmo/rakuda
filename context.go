package rakuda

import (
	"context"
	"log/slog"
	"sync"
)

// contextKey is the type for keys stored in context.
type contextKey string

// Keys for context values.
const (
	loggerKey = contextKey("logger")
)

var logFallbackOnce sync.Once

// NewContextWithLogger returns a new context with the provided Logger.
func NewContextWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// LoggerFromContext retrieves the Logger from the context.
// If no logger is found, it falls back to slog.Default() and logs a warning on the first call.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}

	logFallbackOnce.Do(func() {
		// Use a background context for the warning log because the request context
		// might be canceled, which would prevent the warning from being logged.
		slog.Default().WarnContext(context.Background(), "Logger not found in context, falling back to default logger. This may indicate a misconfiguration.")
	})

	return slog.Default()
}
