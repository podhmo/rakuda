package responder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// testLogger adapts *testing.T to the responder.Logger interface.
type testLogger struct {
	t      *testing.T
	called bool
	msg    string
	args   []any
}

func (l *testLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.t.Helper()
	l.called = true
	l.msg = msg
	l.args = args
}

func TestJSON(t *testing.T) {
	type responseData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name           string
		data           any
		statusCode     int // 0 means default
		wantStatusCode int
		wantBody       string
		wantErrLog     bool
	}{
		{
			name:           "success - 200 OK",
			data:           responseData{Name: "Gopher", Age: 10},
			statusCode:     0, // default
			wantStatusCode: http.StatusOK,
			wantBody:       `{"name":"Gopher","age":10}` + "\n",
		},
		{
			name:           "success - 201 Created",
			data:           responseData{Name: "Gopher", Age: 10},
			statusCode:     http.StatusCreated,
			wantStatusCode: http.StatusCreated,
			wantBody:       `{"name":"Gopher","age":10}` + "\n",
		},
		{
			name:           "success - no content",
			data:           nil,
			statusCode:     http.StatusNoContent,
			wantStatusCode: http.StatusNoContent,
			wantBody:       "",
		},
		{
			name:           "error - json marshal failure",
			data:           make(chan int), // Cannot be marshaled
			statusCode:     http.StatusInternalServerError,
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "",
			wantErrLog:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			logger := &testLogger{t: t}
			req = WithLogger(req, logger)

			if tt.statusCode != 0 {
				req = WithStatusCode(req, tt.statusCode)
			}

			// Act
			JSON(rr, req, tt.data)

			// Assert Status Code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}

			// Assert Body
			if diff := cmp.Diff(tt.wantBody, rr.Body.String()); diff != "" {
				t.Errorf("unexpected body (-want +got):\n%s", diff)
			}

			// Assert Header
			if tt.data != nil {
				wantContentType := "application/json; charset=utf-8"
				if got := rr.Header().Get("Content-Type"); got != wantContentType {
					t.Errorf("wrong Content-Type header: got %q want %q", got, wantContentType)
				}
			}

			// Assert Logger
			if tt.wantErrLog && !logger.called {
				t.Error("expected logger to be called, but it was not")
			}
			if !tt.wantErrLog && logger.called {
				t.Errorf("expected logger not to be called, but it was: msg=%q", logger.msg)
			}
		})
	}
}
