package rakuda

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

// getLogger retrieves the Logger from the context.
func getLogger(ctx context.Context) (Logger, bool) {
	logger, ok := ctx.Value(loggerKey).(Logger)
	return logger, ok && logger != nil
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

// Responder handles writing JSON responses.
type Responder struct {
	// DefaultLogger is used when no logger is found in the request context.
	// If nil, a default slog.Logger is used.
	DefaultLogger Logger
}

// NewResponder creates a new Responder with a default slog logger.
func NewResponder() *Responder {
	return &Responder{
		DefaultLogger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}
}

// JSON marshals the 'data' payload to JSON and writes it to the response.
func (r *Responder) JSON(w http.ResponseWriter, req *http.Request, data any) {
	ctx := req.Context()

	if err := ctx.Err(); err != nil {
		return // Client disconnected
	}

	status := getStatusCode(ctx)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger, ok := getLogger(ctx)
			if !ok {
				logger = r.DefaultLogger
			}
			logger.ErrorContext(ctx, "failed to encode json response", "error", err)
		}
	}
}
