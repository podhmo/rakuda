package bindingparse

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParsers(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		got, err := String("hello")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if want := "hello"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Int", func(t *testing.T) {
		tests := []struct {
			name    string
			input   string
			want    int
			wantErr bool
		}{
			{name: "positive", input: "123", want: 123, wantErr: false},
			{name: "negative", input: "-45", want: -45, wantErr: false},
			{name: "zero", input: "0", want: 0, wantErr: false},
			{name: "invalid", input: "abc", want: 0, wantErr: true},
			{name: "empty", input: "", want: 0, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Int(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Int() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("Int() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("Int64", func(t *testing.T) {
		tests := []struct {
			name    string
			input   string
			want    int64
			wantErr bool
		}{
			{name: "positive", input: "123", want: 123, wantErr: false},
			{name: "large", input: "9223372036854775807", want: math.MaxInt64, wantErr: false},
			{name: "invalid", input: "abc", want: 0, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Int64(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Int64() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("Int64() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("Bool", func(t *testing.T) {
		tests := []struct {
			name    string
			input   string
			want    bool
			wantErr bool
		}{
			{name: "true_1", input: "true", want: true, wantErr: false},
			{name: "true_2", input: "T", want: true, wantErr: false},
			{name: "true_3", input: "1", want: true, wantErr: false},
			{name: "false_1", input: "false", want: false, wantErr: false},
			{name: "false_2", input: "f", want: false, wantErr: false},
			{name: "false_3", input: "0", want: false, wantErr: false},
			{name: "invalid", input: "yes", want: false, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Bool(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Bool() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("Bool() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("Float64", func(t *testing.T) {
		tests := []struct {
			name    string
			input   string
			want    float64
			wantErr bool
		}{
			{name: "positive", input: "123.45", want: 123.45, wantErr: false},
			{name: "negative", input: "-0.5", want: -0.5, wantErr: false},
			{name: "invalid", input: "abc", want: 0, wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Float64(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Float64() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("Float64() mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}

// testValidatable is a test struct that implements the Validator interface.
type testValidatable struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Validate implements the Validator interface.
// It returns an error if the Name is empty or the Value is negative.
func (v *testValidatable) Validate() error {
	if strings.TrimSpace(v.Name) == "" {
		return errors.New("name is required")
	}
	if v.Value < 0 {
		return errors.New("value cannot be negative")
	}
	return nil
}

// parserForTestValidatable is a simple parser for testValidatable, using JSON unmarshaling.
func parserForTestValidatable(s string) (*testValidatable, error) {
	var v testValidatable
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func TestWithValidation(t *testing.T) {
	// Create a new parser that includes the validation step.
	validatedParser := WithValidation(parserForTestValidatable)

	tests := []struct {
		name      string
		input     string
		wantErr   bool
		errString string // Substring to check for in the error message
	}{
		{
			name:      "success",
			input:     `{"name": "test", "value": 10}`,
			wantErr:   false,
			errString: "",
		},
		{
			name:      "parsing_error",
			input:     `{"name": "test", "value": 10`, // Invalid JSON
			wantErr:   true,
			errString: "unexpected end of JSON input",
		},
		{
			name:      "validation_error_name",
			input:     `{"name": "", "value": 10}`,
			wantErr:   true,
			errString: "name is required",
		},
		{
			name:      "validation_error_value",
			input:     `{"name": "test", "value": -5}`,
			wantErr:   true,
			errString: "value cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validatedParser(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("validatedParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("validatedParser() error = %q, want error containing %q", err.Error(), tt.errString)
				}
			}
		})
	}
}
