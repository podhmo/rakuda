package rakuda

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went wrong")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		Recovery(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		expectedContentType := "application/json; charset=utf-8"
		if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
			t.Errorf("expected Content-Type %q, got %q", expectedContentType, contentType)
		}
		expectedBody := `{"error":"Internal Server Error"}` + "\n"
		if rr.Body.String() != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
		}
	})

	t.Run("no panic", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		Recovery(handler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		if rr.Body.String() != "ok" {
			t.Errorf("expected body %q, got %q", "ok", rr.Body.String())
		}
	})
}
