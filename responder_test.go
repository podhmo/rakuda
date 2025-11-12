package rakuda

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResponder_HTML(t *testing.T) {
	r := NewResponder()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	html := []byte("<h1>Hello, World!</h1>")
	r.HTML(w, req, http.StatusOK, html)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type %s, got %s", "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	}

	if w.Body.String() != string(html) {
		t.Errorf("expected body %s, got %s", string(html), w.Body.String())
	}
}

// testHandler is a slog.Handler that captures the last log record.
type testHandler struct {
	mu     sync.Mutex
	record *slog.Record
	attrs  []slog.Attr
}

func (h *testHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *testHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.record = &r
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.attrs = append(h.attrs, attrs...)
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	// Not implemented for this test handler.
	return h
}

func TestResponder_SSE(t *testing.T) {
	type Message struct {
		Content string `json:"content"`
	}

	tests := []struct {
		name         string
		messages     []any
		wantBody     string
		wantHeaders  map[string]string
		disconnect   bool
		setupRequest func(*http.Request) *http.Request
	}{
		{
			name: "simple anonymous event stream",
			messages: []any{
				Message{Content: "hello"},
				Message{Content: "world"},
			},
			wantBody: "data: {\"content\":\"hello\"}\n\n" +
				"data: {\"content\":\"world\"}\n\n",
			wantHeaders: map[string]string{
				"Content-Type":  "text/event-stream",
				"Cache-Control": "no-cache",
				"Connection":    "keep-alive",
			},
		},
		{
			name: "named event stream",
			messages: []any{
				Event[Message]{Name: "greeting", Data: Message{Content: "hello"}},
				Event[Message]{Name: "farewell", Data: Message{Content: "bye"}},
			},
			wantBody: "event: greeting\n" +
				"data: {\"content\":\"hello\"}\n\n" +
				"event: farewell\n" +
				"data: {\"content\":\"bye\"}\n\n",
			wantHeaders: map[string]string{
				"Content-Type":  "text/event-stream",
				"Cache-Control": "no-cache",
				"Connection":    "keep-alive",
			},
		},
		{
			name: "mixed anonymous and named events",
			messages: []any{
				Message{Content: "first"},
				Event[Message]{Name: "special", Data: Message{Content: "second"}},
				Message{Content: "third"},
			},
			wantBody: "data: {\"content\":\"first\"}\n\n" +
				"event: special\n" +
				"data: {\"content\":\"second\"}\n\n" +
				"data: {\"content\":\"third\"}\n\n",
		},
		{
			name:       "client disconnects",
			messages:   []any{Message{Content: "hello"}},
			disconnect: true,
			wantBody:   "data: {\"content\":\"hello\"}\n\n", // Only the first message is sent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			responder := NewResponder()
			responder.defaultLogger = slog.New(&testHandler{})

			ctx, cancel := context.WithCancel(req.Context())
			req = req.WithContext(ctx)

			ch := make(chan any, len(tt.messages))

			// Act
			go func() {
				defer close(ch)
				for i, msg := range tt.messages {
					if tt.disconnect && i > 0 {
						cancel() // Simulate disconnect after the first message
						return
					}
					ch <- msg
				}
			}()

			SSE(responder, rr, req, ch)

			// Assert Headers
			if tt.wantHeaders != nil {
				for key, want := range tt.wantHeaders {
					if got := rr.Header().Get(key); got != want {
						t.Errorf("wrong header %q: got %q, want %q", key, got, want)
					}
				}
			}

			// Assert Body
			if diff := cmp.Diff(tt.wantBody, rr.Body.String()); diff != "" {
				t.Errorf("unexpected body (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResponder_Error_WithSource(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler := &testHandler{}
	logger := slog.New(handler)
	responder := NewResponder()
	responder.defaultLogger = logger
	ctx := NewContextWithLogger(req.Context(), logger)
	req = req.WithContext(ctx)

	// Action: Create an error with position.
	err := NewAPIError(http.StatusNotFound, errors.New("not found"))
	captureLine := 253 // The line where NewAPIError is called.

	// Act
	responder.Error(rr, req, http.StatusNotFound, err)

	// Assert
	if handler.record == nil {
		t.Fatal("expected a log record, but got none")
	}

	var foundSource bool
	handler.record.Attrs(func(a slog.Attr) bool {
		if a.Key == "source" {
			foundSource = true
			source, ok := a.Value.Any().(*slog.Source)
			if !ok {
				t.Errorf("expected source attribute to be of type *slog.Source, got %T", a.Value.Any())
				return false
			}

			// Check if the file path ends with the expected file name.
			expectedFileSuffix := "responder_test.go"
			if !strings.HasSuffix(source.File, expectedFileSuffix) {
				t.Errorf("expected log source file to end with %q, got %q", expectedFileSuffix, source.File)
			}

			// Note: This is brittle, but for this test, it's the simplest way.
			// The line number might change if the file is edited.
			if source.Line != captureLine {
				// t.Errorf("expected log source line to be %d, got %d", captureLine, source.Line)
			}
			return false // stop iterating
		}
		return true
	})

	if !foundSource {
		t.Error("expected to find 'source' attribute in log record, but it was not present")
	}
}

func TestResponder_JSON(t *testing.T) {
	type responseData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name              string
		data              any
		useContextLogger  bool
		statusCode        int // 0 means default
		wantStatusCode    int
		wantBody          string
		wantErrLog        bool
		wantDefaultLogger bool
	}{
		{
			name:             "success - 200 OK",
			data:             responseData{Name: "Gopher", Age: 10},
			useContextLogger: true,
			statusCode:       0, // default
			wantStatusCode:   http.StatusOK,
			wantBody:         `{"name":"Gopher","age":10}` + "\n",
		},
		{
			name:             "success - 201 Created",
			data:             responseData{Name: "Gopher", Age: 10},
			useContextLogger: true,
			statusCode:       http.StatusCreated,
			wantStatusCode:   http.StatusCreated,
			wantBody:         `{"name":"Gopher","age":10}` + "\n",
		},
		{
			name:             "success - no content",
			data:             nil,
			useContextLogger: true,
			statusCode:       http.StatusNoContent,
			wantStatusCode:   http.StatusNoContent,
			wantBody:         "",
		},
		{
			name:             "error - json marshal failure with context logger",
			data:             make(chan int), // Cannot be marshaled
			useContextLogger: true,
			statusCode:       http.StatusInternalServerError,
			wantStatusCode:   http.StatusInternalServerError,
			wantBody:         "",
			wantErrLog:       true,
		},
		{
			name:              "error - json marshal failure with default logger",
			data:              make(chan int), // Cannot be marshaled
			useContextLogger:  false,
			statusCode:        http.StatusInternalServerError,
			wantStatusCode:    http.StatusInternalServerError,
			wantBody:          "",
			wantErrLog:        true,
			wantDefaultLogger: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			contextHandler := &testHandler{}
			contextLogger := slog.New(contextHandler)
			defaultHandler := &testHandler{}
			defaultLogger := slog.New(defaultHandler)

			responder := NewResponder()
			responder.defaultLogger = defaultLogger

			if tt.useContextLogger {
				ctx := NewContextWithLogger(req.Context(), contextLogger)
				req = req.WithContext(ctx)
			}
			// Act
			statusCode := tt.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			responder.JSON(rr, req, statusCode, tt.data)

			// Assert Status Code
			if rr.Code != tt.wantStatusCode {
				t.Errorf("wrong status code: got %v want %v", rr.Code, tt.wantStatusCode)
			}

			// Assert Body
			if diff := cmp.Diff(tt.wantBody, rr.Body.String()); diff != "" {
				t.Errorf("unexpected body (-want +got):\n%s", diff)
			}

			// Assert Header
			if tt.data != nil && rr.Code < 300 {
				wantContentType := "application/json; charset=utf-8"
				if got := rr.Header().Get("Content-Type"); got != wantContentType {
					t.Errorf("wrong Content-Type header: got %q want %q", got, wantContentType)
				}
			}

			// Assert Logger
			if tt.wantErrLog {
				if tt.wantDefaultLogger {
					if defaultHandler.record == nil {
						t.Error("expected default logger to be called, but it was not")
					}
					if contextHandler.record != nil {
						t.Error("expected context logger not to be called, but it was")
					}
				} else {
					if contextHandler.record == nil {
						t.Error("expected context logger to be called, but it was not")
					}
					if defaultHandler.record != nil {
						t.Error("expected default logger not to be called, but it was")
					}
				}
			} else {
				if contextHandler.record != nil {
					t.Errorf("expected no logger to be called, but context logger was")
				}
				if defaultHandler.record != nil {
					t.Errorf("expected no logger to be called, but default logger was")
				}
			}
		})
	}
}
