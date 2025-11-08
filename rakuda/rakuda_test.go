package rakuda

import (
	"net/http"
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

	tests := []struct {
		name           string
		register       func(*Builder)
		expectedMethod string
	}{
		{"Get", func(b *Builder) { b.Get(handler) }, http.MethodGet},
		{"Post", func(b *Builder) { b.Post(handler) }, http.MethodPost},
		{"Put", func(b *Builder) { b.Put(handler) }, http.MethodPut},
		{"Delete", func(b *Builder) { b.Delete(handler) }, http.MethodDelete},
		{"Patch", func(b *Builder) { b.Patch(handler) }, http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			tt.register(b)

			expected := []handlerRegistration{
				{
					method:  tt.expectedMethod,
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
