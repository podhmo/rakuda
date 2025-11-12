package rakuda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/podhmo/rakuda/binding"
)

// Responder handles writing JSON responses.
type Responder struct {
	// defaultLogger is used when no logger is found in the request context.
	// If nil, a default slog.Logger is used.
	defaultLogger *slog.Logger
}

// NewResponder creates a new Responder with a default slog logger.
func NewResponder() *Responder {
	return &Responder{
		defaultLogger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}
}

// Logger returns the logger from the context if it exists, otherwise it returns the default logger.
func (r *Responder) Logger(ctx context.Context) *slog.Logger {
	if logger, ok := LoggerFromContext(ctx); ok {
		return logger
	}
	return r.defaultLogger
}

// Error sends a JSON error response.
// It logs errors only under specific conditions:
// - If the status code is >= 500.
// - If the logger's level is Debug or lower.
// For 5xx errors, it sends a generic message to the client.
func (r *Responder) Error(w http.ResponseWriter, req *http.Request, statusCode int, err error) {
	ctx := req.Context()
	logger := r.Logger(ctx)

	if statusCode >= http.StatusInternalServerError || logger.Enabled(ctx, slog.LevelDebug) {
		attrs := []slog.Attr{
			slog.Int("status", statusCode),
			slog.String("error", err.Error()),
		}

		var apiErr *APIError
		if errors.As(err, &apiErr) {
			if pc := apiErr.PC(); pc != 0 {
				fs := runtime.CallersFrames([]uintptr{pc})
				f, _ := fs.Next()
				if f.File != "" {
					source := &slog.Source{
						File:     f.File,
						Line:     f.Line,
						Function: f.Function,
					}
					attrs = append(attrs, slog.Any("source", source))
				}
			}
		}
		logger.LogAttrs(ctx, slog.LevelError, err.Error(), attrs...)
	}

	var vErrs *binding.ValidationErrors
	if errors.As(err, &vErrs) {
		r.JSON(w, req, statusCode, vErrs)
		return
	}

	errMsg := err.Error()
	if statusCode >= http.StatusInternalServerError {
		// Do not expose internal error details to the client
		errMsg = "Internal Server Error"
	}

	r.JSON(w, req, statusCode, map[string]string{"error": errMsg})
}

// JSON marshals the 'data' payload to JSON and writes it to the response.
func (r *Responder) JSON(w http.ResponseWriter, req *http.Request, statusCode int, data any) {
	ctx := req.Context()

	if err := ctx.Err(); err != nil {
		return // Client disconnected
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger := r.Logger(ctx)
			logger.ErrorContext(ctx, "failed to encode json response", "error", err)
		}
	}
}

// Redirect performs an HTTP redirect.
func (r *Responder) Redirect(w http.ResponseWriter, req *http.Request, url string, code int) {
	http.Redirect(w, req, url, code)
}

// HTML sends an HTML response to the client. This method is intended for use in
// standard http.Handlers, not with Lift, which is designed for JSON APIs.
func (r *Responder) HTML(w http.ResponseWriter, req *http.Request, code int, html []byte) {
	ctx := req.Context()

	if err := ctx.Err(); err != nil {
		return // Client disconnected
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	if _, err := w.Write(html); err != nil {
		logger := r.Logger(ctx)
		logger.ErrorContext(ctx, "failed to write html response", "error", err)
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
		err := fmt.Errorf("Streaming unsupported")
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
