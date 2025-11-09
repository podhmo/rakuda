package rakuda

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

// APIError is an error type that includes an HTTP status code.
type APIError struct {
	err    error
	status int
}

// NewAPIError creates a new APIError.
func NewAPIError(statusCode int, err error) error {
	return &APIError{status: statusCode, err: err}
}

// NewAPIErrorf creates a new APIError with a formatted message.
func NewAPIErrorf(statusCode int, format string, args ...any) *APIError {
	return &APIError{status: statusCode, err: fmt.Errorf(format, args...)}
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return e.err.Error()
}

// StatusCode returns the HTTP status code.
func (e *APIError) StatusCode() int {
	return e.status
}

// Unwrap supports errors.Is and errors.As.
func (e *APIError) Unwrap() error {
	return e.err
}

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
	// defaultLogger is used when no logger is found in the request context.
	// If nil, a default slog.Logger is used.
	defaultLogger Logger
}

// NewResponder creates a new Responder with a default slog logger.
func NewResponder() *Responder {
	return &Responder{
		defaultLogger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}
}

// Logger returns the logger from the context if it exists, otherwise it returns the default logger.
func (r *Responder) Logger(ctx context.Context) Logger {
	if logger, ok := getLogger(ctx); ok {
		return logger
	}
	return r.defaultLogger
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
			logger := r.Logger(ctx)
			logger.ErrorContext(ctx, "failed to encode json response", "error", err)
		}
	}
}

// eventer is a private interface used to extract name and data from a generic Event.
type eventer interface {
	eventName() string
	eventData() any
}

// Event represents a single Server-Sent Event.
type Event[T any] struct {
	// Name is the event name. If empty, it will be omitted.
	Name string
	// Data is the payload for the event.
	Data T
}

// eventName implements the eventer interface.
func (e Event[T]) eventName() string {
	return e.Name
}

// eventData implements the eventer interface.
func (e Event[T]) eventData() any {
	return e.Data
}

// SSE streams data from a channel to the client using the Server-Sent Events protocol.
// It sets the appropriate headers and handles the event stream formatting.
// The channel element type T can be any marshalable type. If T is of type Event[U]
// or *Event[U], it will be treated as a named event.
func SSE[T any](responder *Responder, w http.ResponseWriter, req *http.Request, ch <-chan T) {
	ctx := req.Context()
	logger := responder.Logger(ctx)

	flusher, ok := w.(http.Flusher)
	if !ok {
		err := NewAPIErrorf(http.StatusInternalServerError, "Streaming unsupported")
		http.Error(w, err.Error(), err.StatusCode())
		logger.ErrorContext(ctx, "ResponseWriter does not support flushing", "error", err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
		case msg, ok := <-ch:
			if !ok {
				// Channel closed
				return
			}

			var eventName string
			var dataPayload any = msg

			// Check if the message is an eventer (i.e., an Event or *Event).
			if ev, ok := any(msg).(eventer); ok {
				eventName = ev.eventName()
				dataPayload = ev.eventData()
			}

			// Marshal the data payload to JSON.
			jsonData, err := json.Marshal(dataPayload)
			if err != nil {
				logger.ErrorContext(ctx, "failed to marshal SSE data to JSON", "error", err, "data", dataPayload)
				continue // Skip this message
			}

			if eventName != "" {
				if _, err := fmt.Fprintf(w, "event: %s\n", eventName); err != nil {
					logger.ErrorContext(ctx, "failed to write SSE event name", "error", err)
					return
				}
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
				logger.ErrorContext(ctx, "failed to write SSE data", "error", err)
				return
			}

			flusher.Flush()
		}
	}
}
