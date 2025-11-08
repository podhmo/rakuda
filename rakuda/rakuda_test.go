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

			if len(b.node.actions) != 1 {
				t.Fatalf("expected 1 action, got %d", len(b.node.actions))
			}
			ha, ok := b.node.actions[0].(handlerAction)
			if !ok {
				t.Fatalf("expected handlerAction, got %T", b.node.actions[0])
			}

			if ha.method != tt.expectedMethod {
				t.Errorf("method mismatch: got %q, want %q", ha.method, tt.expectedMethod)
			}
			if ha.pattern != pattern {
				t.Errorf("pattern mismatch: got %q, want %q", ha.pattern, pattern)
			}
			// Compare function pointers.
			got := reflect.ValueOf(ha.handler).Pointer()
			want := reflect.ValueOf(handler).Pointer()
			if got != want {
				t.Errorf("handler function pointer mismatch: got %v, want %v", got, want)
			}
		})
	}
}

func TestOrderIndependence(t *testing.T) {
	// Define handlers and middlewares
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler")) })
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Middleware", "mw")
			next.ServeHTTP(w, r)
		})
	}

	// Builder 1: Register handler then middleware
	b1 := NewBuilder()
	b1.Route("/api", func(b *Builder) {
		b.Get("/handler", handler)
		b.Use(mw)
	})
	router1 := b1.Build()

	// Builder 2: Register middleware then handler
	b2 := NewBuilder()
	b2.Route("/api", func(b *Builder) {
		b.Use(mw)
		b.Get("/handler", handler)
	})
	router2 := b2.Build()

	// --- Verification ---
	path := "/api/handler"
	req := httptest.NewRequest(http.MethodGet, path, nil)

	// Test router 1
	rr1 := httptest.NewRecorder()
	router1.ServeHTTP(rr1, req)

	// Test router 2
	rr2 := httptest.NewRecorder()
	router2.ServeHTTP(rr2, req)

	// Compare responses
	if rr1.Code != rr2.Code {
		t.Errorf("HTTP Status code mismatch: router1=%d, router2=%d", rr1.Code, rr2.Code)
	}
	if diff := cmp.Diff(rr1.Body.String(), rr2.Body.String()); diff != "" {
		t.Errorf("HTTP Body mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(rr1.Header(), rr2.Header()); diff != "" {
		t.Errorf("HTTP Header mismatch (-want +got):\n%s", diff)
	}

	// Explicitly check header for sanity
	wantHeaders := http.Header{}
	wantHeaders.Add("X-Middleware", "mw")
	opts := cmpopts.IgnoreMapEntries(func(key string, val []string) bool {
		return key == "Content-Type"
	})
	if diff := cmp.Diff(wantHeaders, rr1.Header(), opts); diff != "" {
		t.Errorf("Router 1 headers mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantHeaders, rr2.Header(), opts); diff != "" {
		t.Errorf("Router 2 headers mismatch (-want +got):\n%s", diff)
	}
}

func TestGroup(t *testing.T) {
	// Define handlers and middlewares
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler1")) })
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler2")) })
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
	b.Group(func(b *Builder) {
		b.Use(mw1)
		b.Get("/handler1", handler1)
		b.Group(func(b *Builder) {
			b.Use(mw2)
			b.Get("/handler2", handler2)
		})
	})
	router := b.Build()

	// --- Verification ---
	// Test handler1
	req1 := httptest.NewRequest(http.MethodGet, "/handler1", nil)
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)

	if rr1.Body.String() != "handler1" {
		t.Errorf("handler1 body mismatch: got %q, want %q", rr1.Body.String(), "handler1")
	}
	if rr1.Header().Get("X-Middleware-1") != "mw1" {
		t.Errorf("handler1 X-Middleware-1 mismatch: got %q, want %q", rr1.Header().Get("X-Middleware-1"), "mw1")
	}
	if rr1.Header().Get("X-Middleware-2") != "" {
		t.Errorf("handler1 X-Middleware-2 should be absent, got %q", rr1.Header().Get("X-Middleware-2"))
	}

	// Test handler2
	req2 := httptest.NewRequest(http.MethodGet, "/handler2", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	if rr2.Body.String() != "handler2" {
		t.Errorf("handler2 body mismatch: got %q, want %q", rr2.Body.String(), "handler2")
	}
	if rr2.Header().Get("X-Middleware-1") != "mw1" {
		t.Errorf("handler2 X-Middleware-1 mismatch: got %q, want %q", rr2.Header().Get("X-Middleware-1"), "mw1")
	}
	if rr2.Header().Get("X-Middleware-2") != "mw2" {
		t.Errorf("handler2 X-Middleware-2 mismatch: got %q, want %q", rr2.Header().Get("X-Middleware-2"), "mw2")
	}
}
