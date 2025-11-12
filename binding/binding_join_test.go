package binding_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/podhmo/rakuda"
	"github.com/podhmo/rakuda/binding"
)

// parseInt is a simple parser for converting a string to an int.
var parseInt = func(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseString is a simple parser that returns the input string.
var parseString = func(s string) (string, error) {
	return s, nil
}

type MyParams struct {
	ID    int
	Token string
	Sort  *string
}

func TestBindingJoin(t *testing.T) {
	// Create a test request with invalid parameters to trigger validation errors.
	req := httptest.NewRequest("GET", "/users/invalid-id?sort=name", nil)
	// Missing "X-Auth-Token" header, which is required.

	w := httptest.NewRecorder()
	r := req

	var params MyParams
	b := binding.New(r, r.PathValue)

	err := binding.Join(
		binding.One(b, &params.ID, binding.Path, "id", parseInt, binding.Required),
		binding.One(b, &params.Token, binding.Header, "X-Auth-Token", parseString, binding.Required),
		binding.OnePtr(b, &params.Sort, binding.Query, "sort", parseString, binding.Optional),
	)

	// We expect a ValidationErrors type.
	var validationErrs *binding.ValidationErrors
	if ok := errors.As(err, &validationErrs); !ok {
		t.Fatalf("expected error to be of type *binding.ValidationErrors, but got %T", err)
	}

	// Use the responder to generate the JSON output.
	responder := rakuda.NewResponder()
	responder.Error(w, r, http.StatusBadRequest, err)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, but got %d", http.StatusBadRequest, res.StatusCode)
	}

	// Define the expected JSON structure.
	// Note: We're not simulating a real path value call, so the 'id' binding will also fail.
	// This is okay, as it lets us test multiple errors at once.
	// Since there's no actual path routing, r.PathValue("id") returns "", which triggers a "required" error.
	expected := `
	{
		"errors": [
			{
				"source": "path",
				"key": "id",
				"value": null,
				"message": "required parameter is missing"
			},
			{
				"source": "header",
				"key": "X-Auth-Token",
				"value": null,
				"message": "required parameter is missing"
			}
		]
	}`

	var gotJSON, expectedJSON any
	if err := json.NewDecoder(res.Body).Decode(&gotJSON); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		t.Fatalf("failed to unmarshal expected JSON: %v", err)
	}

	if diff := cmp.Diff(expectedJSON, gotJSON); diff != "" {
		t.Errorf("JSON response mismatch (-want +got):\n%s", diff)
	}
}

// TestBindingJoinWithLift simulates a full end-to-end request using the Lift helper.
func TestBindingJoinWithLift(t *testing.T) {
	type GistParams struct {
		ID    int
		Token string
		Sort  *string
	}

	action := func(r *http.Request) (GistParams, error) {
		var params GistParams
		b := binding.New(r, r.PathValue)

		err := binding.Join(
			binding.One(b, &params.ID, binding.Path, "id", parseInt, binding.Required),
			binding.One(b, &params.Token, binding.Header, "X-Auth-Token", parseString, binding.Required),
			binding.OnePtr(b, &params.Sort, binding.Query, "sort", parseString, binding.Optional),
		)
		return params, err
	}

	responder := rakuda.NewResponder()
	handler := rakuda.Lift(responder, action)

	// Create a test request that is missing the required header.
	req := httptest.NewRequest("GET", "/gists/123", nil)
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, but got %d", http.StatusBadRequest, res.StatusCode)
	}

	expected := `
	{
		"errors": [
			{
				"source": "header",
				"key": "X-Auth-Token",
				"value": null,
				"message": "required parameter is missing"
			}
		]
	}`

	var gotJSON, expectedJSON any
	if err := json.NewDecoder(res.Body).Decode(&gotJSON); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		t.Fatalf("failed to unmarshal expected JSON: %v", err)
	}

	if diff := cmp.Diff(expectedJSON, gotJSON); diff != "" {
		t.Errorf("JSON response mismatch (-want +got):\n%s", diff)
	}
}
