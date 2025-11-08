package rakudatest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ResponseAssertion is a function that performs an assertion on an HTTP response.
// It receives the testing object, the original response, and the full response body,
// which has already been read from the response stream.
type ResponseAssertion func(t *testing.T, res *http.Response, body []byte)

// Do executes an HTTP request, checks for a specific status code, runs custom
// assertions on the response, and finally decodes the JSON response body into
// a specified type `T`.
//
// If the actual status code does not match `wantStatusCode`, the test is failed
// with `t.Fatalf`, and the full response body is logged for debugging.
//
// After the status code check, any provided `ResponseAssertion` functions are executed.
//
// If decoding the JSON body fails, the test is also failed with `t.Fatalf`.
func Do[T any](t *testing.T, h http.Handler, req *http.Request, wantStatusCode int, assertions ...ResponseAssertion) T {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("request %s %s: failed to read response body: %v", req.Method, req.URL.Path, err)
	}

	if res.StatusCode != wantStatusCode {
		t.Fatalf("request %s %s: expected status code %d, got %d\nresponse body:\n%s", req.Method, req.URL.Path, wantStatusCode, res.StatusCode, string(body))
	}

	for _, assert := range assertions {
		assert(t, res, body)
	}

	// For 204 No Content, we expect an empty body and return the zero value of T.
	if wantStatusCode == http.StatusNoContent {
		var zero T
		return zero
	}

	var got T
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("request %s %s: failed to decode json response into %T: %v\nresponse body:\n%s", req.Method, req.URL.Path, got, err, string(body))
	}

	return got
}
