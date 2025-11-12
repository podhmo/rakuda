# Design Plan: `binding.Join` for Structured Error Responses

## 1. Motivation

Currently, the `binding` package uses `errors.Join` to combine validation errors. This produces a single, flat error string, which is not ideal for API responses. When a client sends a request with multiple invalid parameters, the server should return a structured error response that clearly identifies each problem.

This proposal outlines a new `binding.Join` function and a set of structured error types that will integrate with the `lift` and `responder` components to generate detailed, user-friendly JSON error responses.

## 2. Proposed Changes

### 2.1. New Error Types in `binding` package

We will introduce two new error types in `binding/binding.go`.

**`binding.Error`**: Represents a single binding error.

```go
// binding/binding.go

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
```

**`binding.ValidationErrors`**: A container for multiple `binding.Error` instances.

```go
// binding/binding.go

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
```

### 2.2. New `binding.Join` Function

This function will replace the direct usage of `errors.Join`. It will filter `nil` errors and collect all `binding.Error` instances into a single `ValidationErrors` struct.

```go
// binding/binding.go

// Join collects binding errors into a single ValidationErrors instance.
// It filters out nil errors. If no errors are found, it returns nil.
func Join(errs ...error) error {
	var validationErrs []*Error
	for _, err := range errs {
		if err == nil {
			continue
		}
		// Check if it's a ValidationErrors and unpack its inner errors
		if vErrs, ok := err.(*ValidationErrors); ok {
			validationErrs = append(validationErrs, vErrs.Errors...)
		} else if bErr, ok := err.(*Error); ok {
			// It's a single binding error
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
```

### 2.3. Update Core Binding Functions

The functions `One`, `OnePtr`, `Slice`, and `SlicePtr` will be updated to return `*binding.Error` on failure.

```go
// binding/binding.go (example for `One`)

func One[T any](b *Binding, dest *T, source Source, key string, parse Parser[T], req Requirement) error {
	valStr, ok := b.Lookup(source, key)
	if !ok {
		if req == Required {
			return &Error{ // <-- Return structured error
				Source: source,
				Key:    key,
				Err:    errors.New("required parameter is missing"),
			}
		}
		return nil
	}

	val, err := parse(valStr)
	if err != nil {
		return &Error{ // <-- Return structured error
			Source: source,
			Key:    key,
			Value:  valStr,
			Err:    err,
		}
	}
	*dest = val
	return nil
}
```

### 2.4. Integration with `responder`

The `responder.Error` method will be updated to recognize `*binding.ValidationErrors`. When it sees this error type, it will marshal it directly to JSON instead of wrapping it in the standard `{"error": "message"}` structure.

```go
// responder.go

func (r *Responder) Error(w http.ResponseWriter, req *http.Request, statusCode int, err error) {
	// ... (logging remains the same)

	var vErrs *ValidationErrors
	if errors.As(err, &vErrs) {
		// For validation errors, send the detailed structured response.
		r.JSON(w, req, statusCode, vErrs)
		return
	}

	errMsg := err.Error()
	if statusCode >= http.StatusInternalServerError {
		errMsg = "Internal Server Error"
	}

	r.JSON(w, req, statusCode, map[string]string{"error": errMsg})
}
```

The `lift` function does not require changes because `ValidationErrors` implements the `StatusCode() int` interface, which `lift` already understands.

## 3. Example Usage

The handler code will now use `binding.Join`:

```go
// examples/simple-rest-api/main.go

func handler(w http.ResponseWriter, r *http.Request) {
	b := binding.New(r, r.PathValue)
	var params MyParams
	if err := binding.Join( // <-- Use binding.Join
		binding.One(b, &params.ID, binding.Path, "id", parseInt, binding.Required),
		binding.One(b, &params.Token, binding.Header, "X-Auth-Token", parseString, binding.Required),
	); err != nil {
		// The error is passed to lift's machinery, which will call responder.Error
		// and generate the structured response.
		// No changes needed here when using Lift.
	}
	// ...
}
```

If a request is missing the `X-Auth-Token` header and has an invalid `id`, the JSON response will be a `400 Bad Request` with the following body:

```json
{
  "errors": [
    {
      "source": "path",
      "key": "id",
      "value": "invalid-id",
      "message": "strconv.Atoi: parsing \"invalid-id\": invalid syntax"
    },
    {
      "source": "header",
      "key": "X-Auth-Token",
      "message": "required parameter is missing"
    }
  ]
}
```

## 4. Testing Strategy

- Add a new test file: `binding/binding_join_test.go`.
- This test will simulate an HTTP request with multiple invalid parameters.
- It will call a handler that uses `binding.Join`.
- It will assert that the HTTP response has a `400` status code and that the JSON body matches the expected structured error format.
- Existing tests for the `binding` package will be updated to reflect the new error types.
