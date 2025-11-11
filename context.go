package rakuda

import (
	"context"
	"io"
	"log/slog"
	"net/http"
)

// contextKey is the type for keys stored in context.
type contextKey string

// Keys for context values.
const (
	loggerKey     = contextKey("logger")
	statusCodeKey = contextKey("status-code")
)

// NewContextWithLogger returns a new context with the provided Logger.
func NewContextWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// LoggerFromContext retrieves the Logger from the context.
// If no logger is found, it returns a disabled logger and false.
func LoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	l, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		// Return a no-op logger if none is found.
		return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelError + 1, // Disable all levels
		})), false
	}
	return l, true
}

// NewContextWithStatusCode returns a new context with the provided HTTP status code.
func NewContextWithStatusCode(ctx context.Context, status int) context.Context {
	return context.WithValue(ctx, statusCodeKey, status)
}

// StatusCodeFromContext retrieves the HTTP status code from the context, or http.StatusOK if not found.
func StatusCodeFromContext(ctx context.Context) int {
	if status, ok := ctx.Value(statusCodeKey).(int); ok {
		return status
	}
	return http.StatusOK
}
