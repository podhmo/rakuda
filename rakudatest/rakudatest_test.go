package rakudatest

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/rakuda"
)

// spyHandler is a handler that retrieves a logger from the context and logs messages.
// It fails the test if the logger is not found or is the default logger.
func spyHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := rakuda.LoggerFromContext(r.Context())
		if logger == slog.Default() {
			t.Error("expected a specific logger from context, but got the default logger")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.DebugContext(r.Context(), "this is a debug message")
		logger.InfoContext(r.Context(), "this is an info message", slog.String("request_id", "xyz-123"))
		logger.WarnContext(r.Context(), "this is a warning message")
		logger.ErrorContext(r.Context(), "this is an error message")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}
}

func TestDo(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/no-content" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 1, "name": "test"}`))
	})

	t.Run("success", func(t *testing.T) {
		type Body struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		want := Body{ID: 1, Name: "test"}
		req := httptest.NewRequest("GET", "/", nil)
		got := Do[Body](t, handler, req, http.StatusOK)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("response body mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("with_assertions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		Do[any](t, handler, req, http.StatusOK, func(t *testing.T, res *http.Response, body []byte) {
			if contentType := res.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("expected Content-Type to be application/json, got %s", contentType)
			}
			if !strings.Contains(string(body), `"name": "test"`) {
				t.Errorf("response body does not contain expected string")
			}
		})
	})

	t.Run("no_content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/no-content", nil)
		got := Do[any](t, handler, req, http.StatusNoContent)
		if got != nil {
			t.Errorf("expected nil body for 204 No Content, got %v", got)
		}
	})
}

func TestDo_WithLogger(t *testing.T) {
	handler := spyHandler(t)

	req := httptest.NewRequest("GET", "/", nil)
	Do[map[string]string](t, handler, req, http.StatusOK)

	// To verify the output, run the test with the verbose flag:
	// go test -v ./...
	// The output should contain the log messages from spyHandler.
}
