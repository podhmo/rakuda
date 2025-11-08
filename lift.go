package rakuda

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError is an error type that includes an HTTP status code.
type APIError struct {
	err    error
	status int
}

// NewAPIError creates a new APIError.
func NewAPIError(statusCode int, err error) *APIError {
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

// Lift converts a function that returns a value and an error into an http.Handler.
//
// The action function has the signature: func(*http.Request) (O, error)
//
// - If the error is nil, the returned value of type O is encoded as a JSON
//   response with a 200 OK status.
// - If the error is not nil:
// - If the error has a StatusCode() int method, its status code is used for the response.
//   - Otherwise, a 500 Internal Server Error is returned.
//   - The error message is returned as a JSON object: {"error": "message"}.
func Lift[O any](responder *Responder, action func(*http.Request) (O, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := action(r)
		if err != nil {
			var sc interface{ StatusCode() int }
			if errors.As(err, &sc) {
				r = WithStatusCode(r, sc.StatusCode())
				responder.JSON(w, r, map[string]string{"error": err.Error()})
				return
			}

			// For internal errors, log the actual error but return a generic message.
			ctx := r.Context()
			logger, ok := getLogger(ctx)
			if !ok || logger == nil {
				logger = responder.DefaultLogger
			}
			logger.ErrorContext(ctx, "internal server error from lifted handler", "error", err)

			r = WithStatusCode(r, http.StatusInternalServerError)
			responder.JSON(w, r, map[string]string{"error": "Internal Server Error"})
			return
		}

		responder.JSON(w, r, data)
	})
}
