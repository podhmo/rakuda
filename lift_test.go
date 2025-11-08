package rakuda

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLift(t *testing.T) {
	type ResponseObject struct {
		Message string `json:"message"`
	}

	tests := []struct {
		name           string
		action         func(*http.Request) (ResponseObject, error)
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "success",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{Message: "hello"}, nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"message":"hello"}`,
		},
		{
			name: "error with status code",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{}, NewAPIError(http.StatusBadRequest, errors.New("invalid input"))
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `{"error":"invalid input"}`,
		},
		{
			name: "error without status code",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{}, errors.New("internal error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `{"error":"Internal Server Error"}`,
		},
		{
			name: "nil data, nil error",
			action: func(r *http.Request) (ResponseObject, error) {
				return ResponseObject{}, nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"message":""}`, // a zero value of the struct
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responder := NewResponder()
			handler := Lift(responder, tt.action)

			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status code %d, got %d", tt.wantStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			// Normalize JSON strings for comparison
			want := strings.TrimSpace(tt.wantBody)
			got := strings.TrimSpace(string(body))

			if diff := cmp.Diff(want, got); diff != "" {
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
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		if len(body) > 0 {
			t.Errorf("expected empty body, but got %q", string(body))
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		responder := NewResponder()
		action := func(r *http.Request) ([]ResponseObject, error) {
			return nil, nil
		}
		handler := Lift(responder, action)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		want := `[]`
		got := strings.TrimSpace(string(body))
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("response body mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		responder := NewResponder()
		action := func(r *http.Request) (map[string]ResponseObject, error) {
			return nil, nil
		}
		handler := Lift(responder, action)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		want := `{}`
		got := strings.TrimSpace(string(body))
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("response body mismatch (-want +got):\n%s", diff)
		}
	})
}
