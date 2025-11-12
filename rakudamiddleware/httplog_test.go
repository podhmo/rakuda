package rakudamiddleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/rakuda"
)

func TestHTTPLog(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		method         string
		path           string
		expectedStatus int
		expectedSize   int
		expectedCType  string
	}{
		{
			name: "GET request with 200 OK",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "hello"}`))
			}),
			method:         http.MethodGet,
			path:           "/test",
			expectedStatus: http.StatusOK,
			expectedSize:   20, // `{"message": "hello"}` is 20 bytes
			expectedCType:  "application/json",
		},
		{
			name: "POST request with 201 Created",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("created"))
			}),
			method:         http.MethodPost,
			path:           "/create",
			expectedStatus: http.StatusCreated,
			expectedSize:   7, // "created" is 7 bytes
			expectedCType:  "text/plain",
		},
		{
			name: "Request with 404 Not Found",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			}),
			method:         http.MethodGet,
			path:           "/notfound",
			expectedStatus: http.StatusNotFound,
			expectedSize:   19, // "Not Found\n" is 10 bytes
			expectedCType:  "text/plain; charset=utf-8",
		},
		{
			name: "Request with no Content-Type",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
			method:         http.MethodDelete,
			path:           "/delete",
			expectedStatus: http.StatusNoContent,
			expectedSize:   0,
			expectedCType:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			// Create a request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			ctx := rakuda.NewContextWithLogger(context.Background(), logger)
			req = req.WithContext(ctx)

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Create the middleware
			middleware := HTTPLog(tt.handler)
			middleware.ServeHTTP(rr, req)

			// Parse the log output
			var logOutput map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logOutput); err != nil {
				t.Fatalf("failed to unmarshal log output: %v", err)
			}

			// Verify the log fields
			if got, want := logOutput["method"], tt.method; got != want {
				t.Errorf("method: got %q, want %q", got, want)
			}
			if got, want := logOutput["path"], tt.path; got != want {
				t.Errorf("path: got %q, want %q", got, want)
			}
			// slog unmarshals numbers as float64
			if got, want := int(logOutput["status"].(float64)), tt.expectedStatus; got != want {
				t.Errorf("status: got %d, want %d", got, want)
			}
			if got, want := int(logOutput["size"].(float64)), tt.expectedSize; got != want {
				t.Errorf("size: got %d, want %d", got, want)
			}
			if got, want := logOutput["content-type"], tt.expectedCType; got != want {
				t.Errorf("content-type: got %q, want %q, diff %s", got, want, cmp.Diff(want, got))
			}
			if _, ok := logOutput["duration"]; !ok {
				t.Error("duration field is missing")
			}
			if _, ok := logOutput["time"]; !ok {
				t.Error("time field is missing")
			}
			if _, ok := logOutput["level"]; !ok {
				t.Error("level field is missing")
			}
			if _, ok := logOutput["msg"]; !ok {
				t.Error("msg field is missing")
			}
		})
	}
}

// TestLogging_DefaultLogger verifies that the middleware uses the default logger when none is in the context.
func TestHTTPLog_DefaultLogger(t *testing.T) {
	// This test doesn't check the output, just that it doesn't panic.
	// A more advanced test could involve redirecting os.Stderr.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked: %v", r)
		}
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	middleware := HTTPLog(handler)
	middleware.ServeHTTP(rr, req)
}

// TestResponseWriter_WriteHeader verifies that the WriteHeader method is called correctly.
func TestResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rr}

	rw.WriteHeader(http.StatusAccepted)

	if got, want := rr.Code, http.StatusAccepted; got != want {
		t.Errorf("WriteHeader: got status %d, want %d", got, want)
	}
	if got, want := rw.status, http.StatusAccepted; got != want {
		t.Errorf("responseWriter.status: got %d, want %d", got, want)
	}
}

// TestResponseWriter_Write verifies that the Write method is called correctly.
func TestResponseWriter_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rr}
	testData := []byte("hello world")

	size, err := rw.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if got, want := size, len(testData); got != want {
		t.Errorf("Write: got size %d, want %d", got, want)
	}
	if got, want := rw.size, len(testData); got != want {
		t.Errorf("responseWriter.size: got %d, want %d", got, want)
	}
	if diff := cmp.Diff(rr.Body.Bytes(), testData); diff != "" {
		t.Errorf("response body mismatch (-want +got):\n%s", diff)
	}
}
