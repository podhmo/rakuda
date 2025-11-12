package rakuda

import (
	"errors"
	"net/http"
	"reflect"
)

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
				http.Redirect(w, r, redirectErr.URL, code)
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
