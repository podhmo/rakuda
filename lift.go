package rakuda

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
)

// APIError is an error type that includes an HTTP status code.
type APIError struct {
	err    error
	status int
	pc     uintptr // program counter
}

// NewAPIError creates a new APIError, capturing the caller's position.
// The default depth is 2, which points to the caller of NewAPIError.
func NewAPIError(statusCode int, err error) *APIError {
	return NewAPIErrorWithDepth(statusCode, err, 2)
}

// NewAPIErrorf creates a new APIError with a formatted message.
// The default depth is 2, which points to the caller of NewAPIErrorf.
func NewAPIErrorf(statusCode int, format string, args ...any) *APIError {
	return NewAPIErrorWithDepth(statusCode, fmt.Errorf(format, args...), 2)
}

// NewAPIErrorWithDepth creates a new APIError with a specific call stack depth.
func NewAPIErrorWithDepth(statusCode int, err error, depth int) *APIError {
	pc, _, _, _ := runtime.Caller(depth)
	return &APIError{
		status: statusCode, err: err, pc: pc,
	}
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return e.err.Error()
}

// StatusCode returns the HTTP status code.
func (e *APIError) StatusCode() int {
	return e.status
}

// PC returns the program counter where the error was created.
func (e *APIError) PC() uintptr {
	return e.pc
}

// Unwrap supports errors.Is and errors.As.
func (e *APIError) Unwrap() error {
	return e.err
}

// RedirectError is a special error type used to signal an HTTP redirect.
// When this error is returned from a handler wrapped by Lift, the Lift
// function will perform the redirect and stop further processing.
type RedirectError struct {
	URL  string
	Code int
}

// Error implements the error interface.
func (e *RedirectError) Error() string {
	return fmt.Sprintf("redirect to %s with code %d", e.URL, e.Code)
}

// Lift converts a function that returns a value and an error into an http.Handler.
//
// The action function has the signature: func(*http.Request) (O, error)
//
//   - If the error is nil, the returned value of type O is encoded as a JSON
//     response with a 200 OK status.
//   - If the error is not nil:
//   - To perform a redirect, return a `*RedirectError`. Lift will handle the
//     redirect and no further response will be written.
//   - If the error has a StatusCode() int method (like `APIError`), its status
//     code is used for the response.
//   - Otherwise, a 500 Internal Server Error is returned.
//   - The error message is returned as a JSON object: {"error": "message"}.
//   - For 5xx errors, the original error is logged, but a generic "Internal Server Error" message
//     is returned to the client to avoid exposing internal details.
//   - If both the returned value and the error are nil, it follows specific rules:
//   - For `nil` maps, it returns `200 OK` with an empty JSON object `{}`.
//   - For `nil` slices, it returns `200 OK` with an empty JSON array `[]`.
//   - For other nillable types (e.g., pointers), it returns `204 No Content`.
func Lift[O any](responder *Responder, action func(*http.Request) (O, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := action(r)
		if err != nil {
			var redirectErr *RedirectError
			if errors.As(err, &redirectErr) {
				code := redirectErr.Code
				if code == 0 {
					code = http.StatusFound
				}
				responder.Redirect(w, r, redirectErr.URL, code)
				return
			}

			var sc interface{ StatusCode() int }
			if errors.As(err, &sc) {
				responder.Error(w, r, sc.StatusCode(), err)
				return
			}
			responder.Error(w, r, http.StatusInternalServerError, err)
			return
		}

		v := reflect.ValueOf(data)
		// Check if the returned value is a nillable type and is nil.
		isNillable := false
		switch v.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface, reflect.Chan, reflect.Func:
			isNillable = true
		}

		if isNillable && v.IsNil() {
			var z O
			typ := reflect.TypeOf(z)

			// For pointer types, we inspect the element type.
			if typ != nil && typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}

			// If the type is still nil (e.g., O is an interface), we can't create
			// a concrete value, so we return No Content.
			if typ == nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			switch typ.Kind() {
			case reflect.Map:
				// For a nil map, return an empty JSON object.
				responder.JSON(w, r, http.StatusOK, reflect.MakeMap(typ).Interface())
				return
			case reflect.Slice:
				// For a nil slice, return an empty JSON array.
				responder.JSON(w, r, http.StatusOK, reflect.MakeSlice(typ, 0, 0).Interface())
				return
			default:
				// For other nil types (pointers, interfaces, etc.), return No Content.
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		// Check if the returned data itself specifies a status code.
		statusCode := http.StatusOK
		if sc, ok := any(data).(interface{ StatusCode() int }); ok {
			statusCode = sc.StatusCode()
		}
		responder.JSON(w, r, statusCode, data)
	})
}
