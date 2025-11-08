package rakuda

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	// Helper function to compare two recorders
	assertRecordersEqual := func(t *testing.T, rr1, rr2 *httptest.ResponseRecorder) {
		t.Helper()
		if rr1.Code != rr2.Code {
			t.Errorf("HTTP Status code mismatch: router1=%d, router2=%d", rr1.Code, rr2.Code)
		}
		if diff := cmp.Diff(rr1.Body.String(), rr2.Body.String()); diff != "" {
			t.Errorf("HTTP Body mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(rr1.Header(), rr2.Header()); diff != "" {
			t.Errorf("HTTP Header mismatch (-want +got):\n%s", diff)
		}
	}

	t.Run("Simple", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler")) })
		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Middleware", "mw")
				next.ServeHTTP(w, r)
			})
		}

		b1 := NewBuilder()
		b1.Route("/api", func(b *Builder) {
			b.Get("/handler", handler)
			b.Use(mw)
		})
		router1 := b1.Build()

		b2 := NewBuilder()
		b2.Route("/api", func(b *Builder) {
			b.Use(mw)
			b.Get("/handler", handler)
		})
		router2 := b2.Build()

		req := httptest.NewRequest(http.MethodGet, "/api/handler", nil)
		rr1 := httptest.NewRecorder()
		rr2 := httptest.NewRecorder()
		router1.ServeHTTP(rr1, req)
		router2.ServeHTTP(rr2, req)

		assertRecordersEqual(t, rr1, rr2)
	})

	t.Run("TopLevelMiddleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler")) })
		topLevelMw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Top-Level", "top")
				next.ServeHTTP(w, r)
			})
		}

		b1 := NewBuilder()
		b1.Route("/api", func(b *Builder) {
			b.Get("/handler", handler)
		})
		b1.Use(topLevelMw) // Applied after
		router1 := b1.Build()

		b2 := NewBuilder()
		b2.Use(topLevelMw) // Applied before
		b2.Route("/api", func(b *Builder) {
			b.Get("/handler", handler)
		})
		router2 := b2.Build()

		req := httptest.NewRequest(http.MethodGet, "/api/handler", nil)
		rr1 := httptest.NewRecorder()
		rr2 := httptest.NewRecorder()
		router1.ServeHTTP(rr1, req)
		router2.ServeHTTP(rr2, req)

		assertRecordersEqual(t, rr1, rr2)
		if rr1.Header().Get("X-Top-Level") != "top" {
			t.Errorf("Expected top-level middleware to be applied")
		}
	})

	t.Run("NestedGroups", func(t *testing.T) {
		parentHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("parent")) })
		nestedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("nested")) })
		parentMw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Parent", "parent-mw")
				next.ServeHTTP(w, r)
			})
		}
		nestedMw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Nested", "nested-mw")
				next.ServeHTTP(w, r)
			})
		}

		b1 := NewBuilder()
		b1.Route("/api", func(b *Builder) {
			b.Route("/v1", func(b *Builder) {
				b.Get("/items", nestedHandler)
				b.Use(nestedMw)
			})
			b.Get("/data", parentHandler)
			b.Use(parentMw)
		})
		router1 := b1.Build()

		b2 := NewBuilder()
		b2.Route("/api", func(b *Builder) {
			b.Use(parentMw)
			b.Get("/data", parentHandler)
			b.Route("/v1", func(b *Builder) {
				b.Use(nestedMw)
				b.Get("/items", nestedHandler)
			})
		})
		router2 := b2.Build()

		// Test parent
		reqParent := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		rrParent1 := httptest.NewRecorder()
		rrParent2 := httptest.NewRecorder()
		router1.ServeHTTP(rrParent1, reqParent)
		router2.ServeHTTP(rrParent2, reqParent)
		assertRecordersEqual(t, rrParent1, rrParent2)
		if rrParent1.Header().Get("X-Parent") != "parent-mw" {
			t.Errorf("Expected parent middleware to be applied")
		}
		if rrParent1.Header().Get("X-Nested") != "" {
			t.Errorf("Did not expect nested middleware to be applied")
		}

		// Test nested
		reqNested := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
		rrNested1 := httptest.NewRecorder()
		rrNested2 := httptest.NewRecorder()
		router1.ServeHTTP(rrNested1, reqNested)
		router2.ServeHTTP(rrNested2, reqNested)
		assertRecordersEqual(t, rrNested1, rrNested2)
		if rrNested1.Header().Get("X-Parent") != "parent-mw" {
			t.Errorf("Expected parent middleware to be applied to nested handler")
		}
		if rrNested1.Header().Get("X-Nested") != "nested-mw" {
			t.Errorf("Expected nested middleware to be applied")
		}
	})
}

func TestWalkAndPrintRoutes(t *testing.T) {
	b := NewBuilder()
	nullHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Define a simple route structure
	b.Get("/a", nullHandler)
	b.Post("/b", nullHandler)
	b.Route("/v1", func(b *Builder) {
		b.Get("/users", nullHandler)
		b.Group(func(b *Builder) {
			b.Put("/users/{id}", nullHandler)
		})
	})

	// 1. Test Walk
	var walkedRoutes [][2]string
	b.Walk(func(method, pattern string) {
		walkedRoutes = append(walkedRoutes, [2]string{method, pattern})
	})

	expectedWalk := [][2]string{
		{http.MethodGet, "/a"},
		{http.MethodPost, "/b"},
		{http.MethodGet, "/v1/users"},
		{http.MethodPut, "/v1/users/{id}"},
	}
	if diff := cmp.Diff(expectedWalk, walkedRoutes); diff != "" {
		t.Errorf("Walk() mismatch (-want +got):\n%s", diff)
	}

	// 2. Test PrintRoutes
	var buf strings.Builder
	PrintRoutes(&buf, b)
	got := buf.String()
	want := `
GET   /a
POST  /b
GET   /v1/users
PUT   /v1/users/{id}
`
	// Normalize whitespace for comparison
	normalize := func(s string) string {
		return strings.TrimSpace(strings.ReplaceAll(s, "\t", "  "))
	}

	if diff := cmp.Diff(normalize(want), normalize(got)); diff != "" {
		t.Errorf("PrintRoutes() mismatch (-want +got):\n%s", diff)
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
