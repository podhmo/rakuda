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

func TestBuildOrderIndependent(t *testing.T) {
	// Define handlers
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler1")) })
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("handler2")) })

	// Builder 1: Register in one order
	b1 := NewBuilder()
	b1.Get("/one", handler1)
	b1.Post("/two", handler2)
	router1 := b1.Build()

	// Builder 2: Register in the opposite order
	b2 := NewBuilder()
	b2.Post("/two", handler2)
	b2.Get("/one", handler1)
	router2 := b2.Build()

	// Table of tests
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{http.MethodGet, "/one", "handler1"},
		{http.MethodPost, "/two", "handler2"},
	}

	for _, tt := range tests {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			// Test router 1
			req1 := httptest.NewRequest(tt.method, tt.path, nil)
			rr1 := httptest.NewRecorder()
			router1.ServeHTTP(rr1, req1)
			if got := rr1.Body.String(); got != tt.want {
				t.Errorf("router1 response mismatch: got %q, want %q", got, tt.want)
			}

			// Test router 2
			req2 := httptest.NewRequest(tt.method, tt.path, nil)
			rr2 := httptest.NewRecorder()
			router2.ServeHTTP(rr2, req2)
			if got := rr2.Body.String(); got != tt.want {
				t.Errorf("router2 response mismatch: got %q, want %q", got, tt.want)
			}
		})
	}
}
