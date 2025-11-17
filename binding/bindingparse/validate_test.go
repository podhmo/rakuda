package bindingparse

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

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
