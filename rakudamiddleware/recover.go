package rakudamiddleware

import (
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/podhmo/rakuda"
)

// Recovery is a middleware that recovers from panics, logs the panic, and returns a 500 Internal Server Error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger, ok := rakuda.LoggerFromContext(r.Context())
				if !ok {
					logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
				}
				logger.ErrorContext(r.Context(), "panic recovered", "error", err, "stack", string(debug.Stack()))

				// Use the new Error method for a standardized response
				responder := rakuda.NewResponder()
				responder.Error(w, r, http.StatusInternalServerError, http.ErrAbortHandler) // http.ErrAbortHandler is just a sentinel error
			}
		}()
		next.ServeHTTP(w, r)
	})
}
