package rakuda

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("NewBuilder() returned nil")
	}
	if b.node == nil {
		t.Fatal("NewBuilder().node is nil")
	}
}

func TestRegisterHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	pattern := "/test"

	tests := []struct {
		name           string
		register       func(*Builder)
		expectedMethod string
	}{
		{"Get", func(b *Builder) { b.Get(pattern, handler) }, http.MethodGet},
		{"Post", func(b *Builder) { b.Post(pattern, handler) }, http.MethodPost},
		{"Put", func(b *Builder) { b.Put(pattern, handler) }, http.MethodPut},
		{"Delete", func(b *Builder) { b.Delete(pattern, handler) }, http.MethodDelete},
		{"Patch", func(b *Builder) { b.Patch(pattern, handler) }, http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			tt.register(b)

			expected := []handlerRegistration{
				{
					method:  tt.expectedMethod,
					pattern: pattern,
					handler: handler,
				},
			}

			if diff := cmp.Diff(expected, b.node.handlers, cmp.AllowUnexported(handlerRegistration{}), cmpopts.IgnoreFields(handlerRegistration{}, "handler")); diff != "" {
				t.Errorf("Builder.node.handlers mismatch (-want +got):\n%s", diff)
			}
			if len(b.node.handlers) != 1 {
				t.Fatalf("expected 1 handler, got %d", len(b.node.handlers))
			}
			// Compare function pointers.
			got := reflect.ValueOf(b.node.handlers[0].handler).Pointer()
			want := reflect.ValueOf(handler).Pointer()
			if got != want {
				t.Errorf("handler function pointer mismatch: got %v, want %v", got, want)
			}
		})
	}
}

func TestMiddlewareAndGrouping(t *testing.T) {
	// Define handlers
	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("root")) })
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("api")) })
	apiV1Handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("api-v1")) })

	// Define middlewares
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Middleware-1", "mw1")
			next.ServeHTTP(w, r)
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Middleware-2", "mw2")
			next.ServeHTTP(w, r)
		})
	}

	// Build the router
	b := NewBuilder()
	b.Use(mw1)
	b.Get("/root", rootHandler)

	b.Route("/api", func(b *Builder) {
		b.Use(mw2)
		b.Get("/data", apiHandler)
		b.Route("/v1", func(b *Builder) {
			b.Get("/items", apiV1Handler)
		})
	})
	router := b.Build()

	// Table of tests
	tests := []struct {
		path         string
		wantBody     string
		wantHeaders  map[string]string
		absentHeaders []string
	}{
		{
			path:     "/root",
			wantBody: "root",
			wantHeaders: map[string]string{
				"X-Middleware-1": "mw1",
			},
			absentHeaders: []string{"X-Middleware-2"},
		},
		{
			path:     "/api/data",
			wantBody: "api",
			wantHeaders: map[string]string{
				"X-Middleware-1": "mw1",
				"X-Middleware-2": "mw2",
			},
		},
		{
			path:     "/api/v1/items",
			wantBody: "api-v1",
			wantHeaders: map[string]string{
				"X-Middleware-1": "mw1",
				"X-Middleware-2": "mw2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if got := rr.Body.String(); got != tt.wantBody {
				t.Errorf("response body mismatch: got %q, want %q", got, tt.wantBody)
			}

			for key, val := range tt.wantHeaders {
				if got := rr.Header().Get(key); got != val {
					t.Errorf("response header %s mismatch: got %q, want %q", key, got, val)
				}
			}

			for _, key := range tt.absentHeaders {
				if got := rr.Header().Get(key); got != "" {
					t.Errorf("unexpected header %s: got %q", key, got)
				}
			}
		})
	}
}
