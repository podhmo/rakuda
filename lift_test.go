package rakuda

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLift_Redirect(t *testing.T) {
	responder := NewResponder()

	t.Run("with code", func(t *testing.T) {
		action := func(r *http.Request) (any, error) {
			return nil, &RedirectError{URL: "/redirect", Code: http.StatusFound}
		}

		handler := Lift(responder, action)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
		}

		if w.Header().Get("Location") != "/redirect" {
			t.Errorf("expected Location %s, got %s", "/redirect", w.Header().Get("Location"))
		}
	})

	t.Run("without code", func(t *testing.T) {
		action := func(r *http.Request) (any, error) {
			return nil, &RedirectError{URL: "/redirect"}
		}

		handler := Lift(responder, action)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("expected status %d, got %d", http.StatusFound, w.Code)
		}

		if w.Header().Get("Location") != "/redirect" {
			t.Errorf("expected Location %s, got %s", "/redirect", w.Header().Get("Location"))
		}
	})
}
