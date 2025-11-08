package responder

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

// Logger defines the minimal interface for a structured error logger.
// It is compatible with the slog.Logger and can be easily implemented
// by wrappers around other loggers or for testing purposes.
type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}

// defaultLogger is the default logger used when no other logger is specified in the context.
var defaultLogger Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

// contextKey is a private type to prevent collisions with other packages' context keys.
type contextKey string

const (
	loggerKey     contextKey = "logger"
	statusCodeKey contextKey = "status-code"
)

// WithLogger returns a new request with the provided Logger stored in its context.
// This should typically be called once by a middleware at the top level.
func WithLogger(r *http.Request, logger Logger) *http.Request {
	ctx := context.WithValue(r.Context(), loggerKey, logger)
	return r.WithContext(ctx)
}

// getLogger retrieves the Logger from the context, or a default slog logger if not found.
func getLogger(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok && logger != nil {
		return logger
	}
	return defaultLogger
}

// WithStatusCode returns a new request with the provided HTTP status code
// stored in its context. This can be called by any middleware or handler
// to set or override the status for the final response.
func WithStatusCode(r *http.Request, status int) *http.Request {
	ctx := context.WithValue(r.Context(), statusCodeKey, status)
	return r.WithContext(ctx)
}

// getStatusCode retrieves the HTTP status code from the context, or http.StatusOK if not found.
func getStatusCode(ctx context.Context) int {
	if status, ok := ctx.Value(statusCodeKey).(int); ok {
		return status
	}
	return http.StatusOK
}

// JSON marshals the 'data' payload to JSON and writes it to the response.
//
// It performs the following steps:
// 1. Checks if the request context has been canceled (e.g., client disconnected).
//    If so, it returns immediately to prevent "broken pipe" errors.
// 2. Retrieves the HTTP status code from the request context. If not set,
//    it defaults to http.StatusOK (200).
// 3. Sets the "Content-Type" header to "application/json; charset=utf-8".
// 4. Writes the HTTP status code to the response header.
// 5. If data is not nil, it encodes the data to the response writer.
// 6. If encoding fails, it retrieves the Logger from the context. It logs the
//    error with contextual information.
func JSON(w http.ResponseWriter, req *http.Request, data any) {
	ctx := req.Context()

	if err := ctx.Err(); err != nil {
		return
	}

	status := getStatusCode(ctx)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger := getLogger(ctx)
			logger.ErrorContext(ctx, "failed to encode json response", "error", err)
		}
	}
}
