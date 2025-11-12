package rakuda

import (
	"errors"
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
		router1, err := b1.Build()
		if err != nil {
			t.Fatalf("b1.Build() failed: %v", err)
		}

		b2 := NewBuilder()
		b2.Route("/api", func(b *Builder) {
			b.Use(mw)
			b.Get("/handler", handler)
		})
		router2, err := b2.Build()
		if err != nil {
			t.Fatalf("b2.Build() failed: %v", err)
		}

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
		router1, err := b1.Build()
		if err != nil {
			t.Fatalf("b1.Build() failed: %v", err)
		}

		b2 := NewBuilder()
		b2.Use(topLevelMw) // Applied before
		b2.Route("/api", func(b *Builder) {
			b.Get("/handler", handler)
		})
		router2, err := b2.Build()
		if err != nil {
			t.Fatalf("b2.Build() failed: %v", err)
		}

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
		router1, err := b1.Build()
		if err != nil {
			t.Fatalf("b1.Build() failed: %v", err)
		}

		b2 := NewBuilder()
		b2.Route("/api", func(b *Builder) {
			b.Use(parentMw)
			b.Get("/data", parentHandler)
			b.Route("/v1", func(b *Builder) {
				b.Use(nestedMw)
				b.Get("/items", nestedHandler)
			})
		})
		router2, err := b2.Build()
		if err != nil {
			t.Fatalf("b2.Build() failed: %v", err)
		}

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

func TestConflictHandling(t *testing.T) {
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler1")) })
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler2")) })

	t.Run("NoErrorOnNoConflict", func(t *testing.T) {
		b := NewBuilder()
		b.Get("/path1", handler1)
		b.Post("/path1", handler1)
		b.Get("/path2", handler1)

		if _, err := b.Build(); err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("ErrorOnConflictWithCustomFunc", func(t *testing.T) {
		b := NewBuilder(WithOnConflict(func(b *Builder, routeKey string) error {
			return errors.New("custom conflict error")
		}))
		b.Get("/conflict", handler1)
		b.Get("/conflict", handler2)

		_, err := b.Build()
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		expectedErr := "custom conflict error"
		if err.Error() != expectedErr {
			t.Errorf("Error message mismatch:\ngot:  %q\nwant: %q", err.Error(), expectedErr)
		}
	})

	t.Run("DefaultWarningOnConflict", func(t *testing.T) {
		// This test primarily checks that no error is returned with the default behavior.
		// A more robust test would capture log output.
		b := NewBuilder()
		b.Get("/conflict", handler1)
		b.Get("/conflict", handler2)

		if _, err := b.Build(); err != nil {
			t.Errorf("Expected no error for default warn behavior, but got: %v", err)
		}
	})

	t.Run("ConflictInNestedRouteWithError", func(t *testing.T) {
		b := NewBuilder(WithOnConflict(func(b *Builder, routeKey string) error {
			return errors.New("nested conflict")
		}))
		b.Route("/api", func(b *Builder) {
			b.Get("/users", handler1)
		})
		b.Get("/api/users", handler2) // This creates the conflict

		_, err := b.Build()
		if err == nil {
			t.Fatal("Expected an error for nested conflict, but got nil")
		}
		expectedErr := "nested conflict"
		if err.Error() != expectedErr {
			t.Errorf("Error message mismatch for nested conflict:\ngot:  %q\nwant: %q", err.Error(), expectedErr)
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
	router, err := b.Build()
	if err != nil {
		t.Fatalf("b.Build() failed: %v", err)
	}

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

func TestNotFoundHandler(t *testing.T) {
	// Handler for existing routes
	existingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Custom 404 handler
	customNotFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("custom not found"))
	})

	// -- Test Cases --
	t.Run("DefaultNotFound", func(t *testing.T) {
		b := NewBuilder()
		b.Get("/existing", existingHandler)
		router, err := b.Build()
		if err != nil {
			t.Fatalf("b.Build() failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Status code mismatch: got %d, want %d", rr.Code, http.StatusNotFound)
		}
		// The default handler uses a Responder, which adds a newline.
		wantBody := `{"error":"not found"}` + "\n"
		if rr.Body.String() != wantBody {
			t.Errorf("Body mismatch: got %q, want %q", rr.Body.String(), wantBody)
		}
		wantContentType := "application/json; charset=utf-8"
		if got := rr.Header().Get("Content-Type"); got != wantContentType {
			t.Errorf("Content-Type mismatch: got %q, want %q", got, wantContentType)
		}
	})

	t.Run("CustomNotFound", func(t *testing.T) {
		b := NewBuilder()
		b.Get("/existing", existingHandler)
		b.NotFound(customNotFoundHandler)
		router, err := b.Build()
		if err != nil {
			t.Fatalf("b.Build() failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Status code mismatch: got %d, want %d", rr.Code, http.StatusNotFound)
		}
		if rr.Body.String() != "custom not found" {
			t.Errorf("Body mismatch: got %q, want %q", rr.Body.String(), "custom not found")
		}
	})

	t.Run("ExistingRouteUnaffected", func(t *testing.T) {
		b := NewBuilder()
		b.Get("/existing", existingHandler)
		b.NotFound(customNotFoundHandler) // Should not be called
		router, err := b.Build()
		if err != nil {
			t.Fatalf("b.Build() failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/existing", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Status code mismatch: got %d, want %d", rr.Code, http.StatusOK)
		}
		if rr.Body.String() != "ok" {
			t.Errorf("Body mismatch: got %q, want %q", rr.Body.String(), "ok")
		}
	})

	t.Run("RootPathWithNotFound", func(t *testing.T) {
		b := NewBuilder()
		// Register a handler for the root path.
		b.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("root"))
		}))
		b.NotFound(customNotFoundHandler)
		router, err := b.Build()
		if err != nil {
			t.Fatalf("b.Build() failed: %v", err)
		}

		// 1. Test the root path `GET /`
		reqRoot := httptest.NewRequest(http.MethodGet, "/", nil)
		rrRoot := httptest.NewRecorder()
		router.ServeHTTP(rrRoot, reqRoot)

		if rrRoot.Code != http.StatusOK {
			t.Errorf("Root path status mismatch: got %d, want %d", rrRoot.Code, http.StatusOK)
		}
		if rrRoot.Body.String() != "root" {
			t.Errorf("Root path body mismatch: got %q, want %q", rrRoot.Body.String(), "root")
		}

		// 2. Test a non-existent path `/not-found`
		reqNotFound := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		rrNotFound := httptest.NewRecorder()
		router.ServeHTTP(rrNotFound, reqNotFound)

		if rrNotFound.Code != http.StatusNotFound {
			t.Errorf("Not found status mismatch: got %d, want %d", rrNotFound.Code, http.StatusNotFound)
		}
		if rrNotFound.Body.String() != "custom not found" {
			t.Errorf("Not found body mismatch: got %q, want %q", rrNotFound.Body.String(), "custom not found")
		}
	})
}
