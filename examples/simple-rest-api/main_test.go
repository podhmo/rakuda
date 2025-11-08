package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAPI(t *testing.T) {
	handler := newRouter()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	tests := []struct {
		name       string
		path       string
		want       string
		statusCode int
	}{
		{
			name:       "root",
			path:       "/",
			want:       `{"message":"hello world"}` + "\n",
			statusCode: http.StatusOK,
		},
		{
			name:       "hello",
			path:       "/hello/Jules",
			want:       `{"message":"hello Jules"}` + "\n",
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := http.Get(ts.URL + tt.path)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.statusCode {
				t.Errorf("got status code %d, want %d", res.StatusCode, tt.statusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, string(b)); diff != "" {
				t.Errorf("response body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
