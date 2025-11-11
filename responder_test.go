package rakuda

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// testLogger adapts *testing.T to the responder.Logger interface.
type testLogger struct {
	t      *testing.T
	called bool
	msg    string
	args   []any
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
			responder.defaultLogger = &testLogger{t: t}

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

func (l *testLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.t.Helper()
	l.called = true
	l.msg = msg
	l.args = args
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

			contextLogger := &testLogger{t: t, msg: "context"}
			defaultLogger := &testLogger{t: t, msg: "default"}

			responder := NewResponder()
			responder.defaultLogger = defaultLogger

			if tt.useContextLogger {
				req = WithLogger(req, contextLogger)
			}
			if tt.statusCode != 0 {
				req = WithStatusCode(req, tt.statusCode)
			}

			// Act
			responder.JSON(rr, req, tt.data)

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
					if !defaultLogger.called {
						t.Error("expected default logger to be called, but it was not")
					}
					if contextLogger.called {
						t.Error("expected context logger not to be called, but it was")
					}
				} else {
					if !contextLogger.called {
						t.Error("expected context logger to be called, but it was not")
					}
					if defaultLogger.called {
						t.Error("expected default logger not to be called, but it was")
					}
				}
			} else {
				if contextLogger.called {
					t.Errorf("expected no logger to be called, but context logger was")
				}
				if defaultLogger.called {
					t.Errorf("expected no logger to be called, but default logger was")
				}
			}
		})
	}
}
