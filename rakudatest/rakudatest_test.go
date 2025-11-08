package rakudatest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type simpleResponse struct {
	Message string `json:"message"`
}

func TestDo_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Test-Header", "hello-world")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "hello"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Define a ResponseAssertion to check the header
	assertHeader := func(t *testing.T, res *http.Response, body []byte) {
		t.Helper()
		want := "hello-world"
		if got := res.Header.Get("X-Test-Header"); got != want {
			t.Errorf("expected header X-Test-Header to be %q, but got %q", want, got)
		}
	}

	got := Do[simpleResponse](t, handler, req, http.StatusOK, assertHeader)

	want := simpleResponse{Message: "hello"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("decoded response mismatch (-want +got):\n%s", diff)
	}
}

func TestDo_NoContent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := Do[*simpleResponse](t, handler, req, http.StatusNoContent)

	if got != nil {
		t.Errorf("expected a nil response for 204 No Content, but got %+v", got)
	}
}
