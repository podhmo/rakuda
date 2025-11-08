package rakuda

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/rakuda/rakudatest"
)

func TestLift(t *testing.T) {
	type ResponseObject struct {
		Message string `json:"message,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	tests := []struct {
		name           string
		action         func(*http.Request) (ResponseObject, error)
		wantStatusCode int
		wantResponse   ResponseObject
	}{
		{
			name: "success",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{Message: "hello"}, nil
			},
			wantStatusCode: http.StatusOK,
			wantResponse:   ResponseObject{Message: "hello"},
		},
		{
			name: "error with status code",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{}, NewAPIError(http.StatusBadRequest, errors.New("invalid input"))
			},
			wantStatusCode: http.StatusBadRequest,
			wantResponse:   ResponseObject{Error: "invalid input"},
		},
		{
			name: "error without status code",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{}, errors.New("internal error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantResponse:   ResponseObject{Error: "Internal Server Error"},
		},
		{
			name: "nil data, nil error",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{}, nil
			},
			wantStatusCode: http.StatusOK,
			wantResponse:   ResponseObject{}, // zero value of the struct
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responder := NewResponder()
			handler := Lift(responder, tt.action)

			req := httptest.NewRequest("GET", "/", nil)
			got := rakudatest.Do[ResponseObject](t, handler, req, tt.wantStatusCode)

			if diff := cmp.Diff(tt.wantResponse, got); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLift_NilNil(t *testing.T) {
	type ResponseObject struct {
		Message string `json:"message"`
	}

	t.Run("nil pointer", func(t *testing.T) {
		responder := NewResponder()
		action := func(r *http.Request) (*ResponseObject, error) {
			return nil, nil
		}
		handler := Lift(responder, action)

		req := httptest.NewRequest("GET", "/", nil)
		got := rakudatest.Do[*ResponseObject](t, handler, req, http.StatusNoContent)

		if got != nil {
			t.Errorf("expected nil response for 204 No Content, but got %+v", got)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		responder := NewResponder()
		action := func(r *http.Request) ([]ResponseObject, error) {
			return nil, nil
		}
		handler := Lift(responder, action)

		req := httptest.NewRequest("GET", "/", nil)
		got := rakudatest.Do[[]ResponseObject](t, handler, req, http.StatusOK)

		if len(got) != 0 {
			t.Errorf("expected empty slice, but got %v with length %d", got, len(got))
		}
	})

	t.Run("nil map", func(t *testing.T) {
		responder := NewResponder()
		action := func(r *http.Request) (map[string]ResponseObject, error) {
			return nil, nil
		}
		handler := Lift(responder, action)

		req := httptest.NewRequest("GET", "/", nil)
		got := rakudatest.Do[map[string]ResponseObject](t, handler, req, http.StatusOK)

		// A nil map decodes to an empty map `{}`, which is not nil.
		if len(got) != 0 {
			t.Errorf("expected empty map, but got %v", got)
		}
	})
}
