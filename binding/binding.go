// Package binding provides a type-safe, reflect-free, and expression-oriented
// way to bind data from HTTP requests to Go structs.
package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
)

// Error represents a single validation error, providing structured details.
type Error struct {
	Source Source `json:"source"` // e.g., "query", "header"
	Key    string `json:"key"`    // The parameter name (e.g., "id", "sort")
	Value  any    `json:"value"`  // The invalid value that was provided
	Err    error  `json:"-"`      // The underlying error (not exposed in JSON)
}

func (e *Error) Error() string {
	return fmt.Sprintf("source=%s, key=%s, value=%v, err=%v", e.Source, e.Key, e.Value, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// MarshalJSON customizes the JSON output to include a user-friendly message.
func (e *Error) MarshalJSON() ([]byte, error) {
	type Alias Error
	return json.Marshal(&struct {
		Message string `json:"message"`
		*Alias
	}{
		Message: e.Err.Error(),
		Alias:   (*Alias)(e),
	})
}

// ValidationErrors collects multiple binding errors.
type ValidationErrors struct {
	Errors []*Error `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	var b strings.Builder
	b.WriteString("validation failed: ")
	for i, err := range e.Errors {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

// StatusCode returns 400 Bad Request, allowing it to work with the lift handler.
func (e *ValidationErrors) StatusCode() int {
	return http.StatusBadRequest
}

// Join collects binding errors into a single ValidationErrors instance.
// It filters out nil errors. If no errors are found, it returns nil.
func Join(errs ...error) error {
	var validationErrs []*Error
	for _, err := range errs {
		if err == nil {
			continue
		}
		var vErrs *ValidationErrors
		var bErr *Error
		if errors.As(err, &vErrs) {
			validationErrs = append(validationErrs, vErrs.Errors...)
		} else if errors.As(err, &bErr) {
			validationErrs = append(validationErrs, bErr)
		} else {
			// It's some other error type, wrap it for consistency
			validationErrs = append(validationErrs, &Error{Err: err})
		}
	}

	if len(validationErrs) == 0 {
		return nil
	}
	return &ValidationErrors{Errors: validationErrs}
}

// Source represents the source of a value in an HTTP request.
type Source string

const (
	Query  Source = "query"
	Header Source = "header"
	Cookie Source = "cookie"
	Path   Source = "path"
	Form   Source = "form"
)

// Requirement specifies whether a value is required or optional.
type Requirement bool

const (
	Required Requirement = true
	Optional Requirement = false
)

// defaultMaxMemory is the default maximum memory size for parsing multipart forms.
const defaultMaxMemory = 32 << 20 // 32 MB

// Parser is a generic function that parses a string into a value of type T.
// It returns an error if parsing fails.
type Parser[T any] func(string) (T, error)

// Binding holds the context for a binding operation, including the HTTP request
// and a function to retrieve path parameters.
type Binding struct {
	req       *http.Request
	pathValue func(string) string
}

// New creates a new Binding instance from an *http.Request and a function to retrieve path parameters.
// The pathValue function is typically provided by a routing library.
func New(req *http.Request, pathValue func(string) string) *Binding {
	return &Binding{req: req, pathValue: pathValue}
}

// Lookup is an internal method that retrieves a value and its existence from a given source.
func (b *Binding) Lookup(source Source, key string) (string, bool) {
	switch source {
	case Query:
		if b.req.URL.Query().Has(key) {
			return b.req.URL.Query().Get(key), true
		}
		return "", false
	case Header:
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		if vals, ok := b.req.Header[canonicalKey]; ok {
			if len(vals) > 0 {
				return vals[0], true
			}
			return "", true // Key present with empty value
		}
		return "", false
	case Cookie:
		cookie, err := b.req.Cookie(key)
		if err == nil {
			return cookie.Value, true
		}
		return "", false
	case Path:
		if b.pathValue != nil {
			val := b.pathValue(key)
			if val != "" {
				return val, true
			}
		}
		return "", false
	case Form:
		// Calling ParseMultipartForm is safe to call multiple times.
		// According to the Go documentation, after the first call, subsequent calls have no effect.
		// This parsing populates r.PostForm, which contains only values from the request body.
		// We intentionally use r.PostForm instead of r.FormValue to strictly separate
		// form data from URL query parameters, adhering to the package's design of explicit data sources.
		_ = b.req.ParseMultipartForm(defaultMaxMemory)
		if vs, ok := b.req.PostForm[key]; ok && len(vs) > 0 {
			return vs[0], true
		}
		return "", false
	}
	return "", false
}

// valuesFromSource retrieves all values for a given key from the specified source.
func (b *Binding) valuesFromSource(source Source, key string) ([]string, bool) {
	switch source {
	case Query:
		if values, ok := b.req.URL.Query()[key]; ok && len(values) > 0 {
			return values, true
		}
		return nil, false
	case Form:
		// Calling ParseMultipartForm is safe to call multiple times.
		// According to the Go documentation, after the first call, subsequent calls have no effect.
		// This parsing populates r.PostForm, which contains only values from the request body.
		// We intentionally use r.PostForm instead of r.FormValue to strictly separate
		// form data from URL query parameters, adhering to the package's design of explicit data sources.
		_ = b.req.ParseMultipartForm(defaultMaxMemory)
		if values, ok := b.req.PostForm[key]; ok && len(values) > 0 {
			return values, true
		}
		return nil, false
	case Header:
		canonicalKey := textproto.CanonicalMIMEHeaderKey(key)
		if values, ok := b.req.Header[canonicalKey]; ok && len(values) > 0 {
			return values, true
		}
		return nil, false
	case Cookie:
		cookie, err := b.req.Cookie(key)
		if err == nil {
			return []string{cookie.Value}, true
		}
		return nil, false
	case Path:
		if b.pathValue != nil {
			val := b.pathValue(key)
			if val != "" {
				return []string{val}, true
			}
		}
		return nil, false
	}
	return nil, false
}

// One binds a single value of a non-pointer type (e.g., int, string).
func One[T any](b *Binding, dest *T, source Source, key string, parse Parser[T], req Requirement) error {
	valStr, ok := b.Lookup(source, key)
	if !ok {
		if req == Required {
			return &Error{
				Source: source,
				Key:    key,
				Err:    errors.New("required parameter is missing"),
			}
		}
		return nil // Optional and not present is a success.
	}

	val, err := parse(valStr)
	if err != nil {
		return &Error{
			Source: source,
			Key:    key,
			Value:  valStr,
			Err:    err,
		}
	}

	*dest = val
	return nil
}

// OnePtr binds a single value of a pointer type (e.g., *int, *string).
func OnePtr[T any](b *Binding, dest **T, source Source, key string, parse Parser[T], req Requirement) error {
	valStr, ok := b.Lookup(source, key)
	if !ok {
		if req == Required {
			return &Error{
				Source: source,
				Key:    key,
				Err:    errors.New("required parameter is missing"),
			}
		}
		*dest = nil // Optional and not present: set field to nil.
		return nil
	}

	val, err := parse(valStr)
	if err != nil {
		return &Error{
			Source: source,
			Key:    key,
			Value:  valStr,
			Err:    err,
		}
	}

	*dest = &val
	return nil
}

// Slice binds values into a slice of a non-pointer type (e.g., []int, []string).
func Slice[T any](b *Binding, dest *[]T, source Source, key string, parse Parser[T], req Requirement) error {
	rawValues, ok := b.valuesFromSource(source, key)
	if !ok {
		if req == Required {
			return &Error{
				Source: source,
				Key:    key,
				Err:    errors.New("required parameter is missing"),
			}
		}
		*dest = nil
		return nil
	}

	slice := make([]T, 0)
	var errs []error

	for _, valStr := range rawValues {
		itemsStr := strings.Split(valStr, ",")
		for _, itemStr := range itemsStr {
			trimmed := strings.TrimSpace(itemStr)
			val, err := parse(trimmed)
			if err != nil {
				errs = append(errs, &Error{
					Source: source,
					Key:    key,
					Value:  itemStr,
					Err:    err,
				})
				continue
			}
			slice = append(slice, val)
		}
	}

	if len(errs) > 0 {
		*dest = slice
		return Join(errs...)
	}

	*dest = slice
	return nil
}

// SlicePtr binds values into a slice of a pointer type (e.g., []*int, []*string).
func SlicePtr[T any](b *Binding, dest *[]*T, source Source, key string, parse Parser[T], req Requirement) error {
	rawValues, ok := b.valuesFromSource(source, key)
	if !ok {
		if req == Required {
			return &Error{
				Source: source,
				Key:    key,
				Err:    errors.New("required parameter is missing"),
			}
		}
		*dest = nil
		return nil
	}

	slice := make([]*T, 0)
	var errs []error

	for _, valStr := range rawValues {
		itemsStr := strings.Split(valStr, ",")
		for _, itemStr := range itemsStr {
			trimmed := strings.TrimSpace(itemStr)
			val, err := parse(trimmed)
			if err != nil {
				errs = append(errs, &Error{
					Source: source,
					Key:    key,
					Value:  itemStr,
					Err:    err,
				})
				continue
			}
			slice = append(slice, &val)
		}
	}

	if len(errs) > 0 {
		*dest = slice
		return Join(errs...)
	}

	*dest = slice
	return nil
}
