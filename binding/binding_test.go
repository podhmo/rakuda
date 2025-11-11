package binding

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Mock parsers for testing
var (
	parseString Parser[string] = func(s string) (string, error) {
		return s, nil
	}
	parseInt Parser[int] = func(s string) (int, error) {
		return strconv.Atoi(s)
	}
)

func TestOne(t *testing.T) {
	tests := []struct {
		name          string
		req           *http.Request
		pathValue     func(string) string
		dest          any
		source        Source
		key           string
		parser        any
		reqType       Requirement
		wantErr       bool
		expectedValue any
	}{
		{
			name:    "Required Query Param - Found",
			req:     httptest.NewRequest("GET", "/?id=123", nil),
			dest:    new(int),
			source:  Query,
			key:     "id",
			parser:  parseInt,
			reqType: Required,
			wantErr: false,
			expectedValue: 123,
		},
		{
			name:    "Required Query Param - Not Found",
			req:     httptest.NewRequest("GET", "/", nil),
			dest:    new(int),
			source:  Query,
			key:     "id",
			parser:  parseInt,
			reqType: Required,
			wantErr: true,
		},
		{
			name:    "Optional Header - Found",
			req:     httptest.NewRequest("GET", "/", nil),
			dest:    new(string),
			source:  Header,
			key:     "X-Request-ID",
			parser:  parseString,
			reqType: Optional,
			wantErr: false,
			expectedValue: "abc-123",
			// setup in test
		},
		{
			name:    "Optional Header - Not Found",
			req:     httptest.NewRequest("GET", "/", nil),
			dest:    new(string),
			source:  Header,
			key:     "X-Request-ID",
			parser:  parseString,
			reqType: Optional,
			wantErr: false,
			expectedValue: "",
		},
		{
			name:    "Path Parameter - Found",
			req:     httptest.NewRequest("GET", "/users/456", nil),
			pathValue: func(s string) string {
				if s == "userID" {
					return "456"
				}
				return ""
			},
			dest:    new(int),
			source:  Path,
			key:     "userID",
			parser:  parseInt,
			reqType: Required,
			wantErr: false,
			expectedValue: 456,
		},
		{
			name:    "Cookie - Found",
			req:     httptest.NewRequest("GET", "/", nil),
			dest:    new(string),
			source:  Cookie,
			key:     "session-id",
			parser:  parseString,
			reqType: Required,
			wantErr: false,
			expectedValue: "xyz-789",
			// setup in test
		},
		{
			name:    "Parse Error",
			req:     httptest.NewRequest("GET", "/?id=abc", nil),
			dest:    new(int),
			source:  Query,
			key:     "id",
			parser:  parseInt,
			reqType: Required,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Optional Header - Found" {
				tt.req.Header.Set("X-Request-ID", "abc-123")
			}
			if tt.name == "Cookie - Found" {
				tt.req.AddCookie(&http.Cookie{Name: "session-id", Value: "xyz-789"})
			}

			b := New(tt.req, tt.pathValue)
			var err error
			switch dest := tt.dest.(type) {
			case *int:
				err = One(b, dest, tt.source, tt.key, tt.parser.(Parser[int]), tt.reqType)
				if !tt.wantErr && *dest != tt.expectedValue.(int) {
					t.Errorf("One() got = %v, want %v", *dest, tt.expectedValue)
				}
			case *string:
				err = One(b, dest, tt.source, tt.key, tt.parser.(Parser[string]), tt.reqType)
				if !tt.wantErr && *dest != tt.expectedValue.(string) {
					t.Errorf("One() got = %v, want %v", *dest, tt.expectedValue)
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("One() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOnePtr(t *testing.T) {
	t.Run("Optional Query Param - Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?name=jules", nil)
		b := New(req, nil)
		var name *string
		err := OnePtr(b, &name, Query, "name", parseString, Optional)
		if err != nil {
			t.Fatalf("OnePtr() error = %v, want nil", err)
		}
		if name == nil || *name != "jules" {
			t.Errorf("OnePtr() got = %v, want 'jules'", name)
		}
	})

	t.Run("Optional Query Param - Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		b := New(req, nil)
		var name *string
		err := OnePtr(b, &name, Query, "name", parseString, Optional)
		if err != nil {
			t.Fatalf("OnePtr() error = %v, want nil", err)
		}
		if name != nil {
			t.Errorf("OnePtr() got = %v, want nil", name)
		}
	})

	t.Run("Required Query Param - Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		b := New(req, nil)
		var name *string
		err := OnePtr(b, &name, Query, "name", parseString, Required)
		if err == nil {
			t.Fatal("OnePtr() error = nil, want error")
		}
	})
}

func TestSlice(t *testing.T) {
	t.Run("Multiple Query Params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?ids=1&ids=2&ids=3", nil)
		b := New(req, nil)
		var ids []int
		err := Slice(b, &ids, Query, "ids", parseInt, Required)
		if err != nil {
			t.Fatalf("Slice() error = %v, want nil", err)
		}
		expected := []int{1, 2, 3}
		if diff := cmp.Diff(expected, ids); diff != "" {
			t.Errorf("Slice() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Comma-Separated Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Values", "10, 20, 30")
		b := New(req, nil)
		var values []int
		err := Slice(b, &values, Header, "X-Values", parseInt, Required)
		if err != nil {
			t.Fatalf("Slice() error = %v, want nil", err)
		}
		expected := []int{10, 20, 30}
		if diff := cmp.Diff(expected, values); diff != "" {
			t.Errorf("Slice() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Required - Not Found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		b := New(req, nil)
		var ids []int
		err := Slice(b, &ids, Query, "ids", parseInt, Required)
		if err == nil {
			t.Fatal("Slice() error = nil, want error")
		}
	})
}

func TestSlicePtr(t *testing.T) {
	t.Run("Comma-Separated with partial errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Values", "10, twenty, 30")
		b := New(req, nil)

		var values []*int
		err := SlicePtr(b, &values, Header, "X-Values", parseInt, Required)

		if err == nil {
			t.Fatal("SlicePtr() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "twenty") {
			t.Errorf("expected error to contain the failing value, got %v", err)
		}

		// Check that the valid parts were still parsed and appended
		expectedLen := 2
		if len(values) != expectedLen {
			t.Fatalf("SlicePtr() len = %d, want %d", len(values), expectedLen)
		}
		if *values[0] != 10 || *values[1] != 30 {
			t.Errorf("SlicePtr() got partial result %v, want [10, 30]", values)
		}
	})
}
