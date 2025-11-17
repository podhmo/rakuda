package bindingparse

import (
	"math"
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
