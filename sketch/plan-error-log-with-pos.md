> [!NOTE]
> This feature has been implemented.

# Plan: Add caller position to error logs

## Motivation

Currently, when `log/slog` is configured with `AddSource: true`, it logs the file and line number where the log function (e.g., `slog.Error`) was called. In our framework, errors are often handled centrally, for example, within the `responder.Error` method. This means that all error logs point to the same line in `responder.go`, which is not helpful for debugging. We want to log the position where the error was *actually* generated, for example, inside a `lift`-wrapped action handler.

## Goal

Modify the error handling and logging mechanism to include the file and line number of the original caller that created the error.

## Proposal

We will leverage the `runtime` package to capture the program counter (PC) at the time an error is created and then use it later during logging to resolve the file and line number.

### 1. Enhance `APIError`

The `APIError` struct in `lift.go` will be modified to store the program counter.

```go
// in lift.go

type APIError struct {
	err        error
	StatusCode int
	pc         uintptr // program counter
}
```

### 2. Create a PC-aware constructor for `APIError`

We will introduce a new function, `NewAPIError`, that captures the caller's PC. This function will take a `depth` argument to specify which frame in the call stack to capture. The default will be `2` (`NewAPIError` -> its caller).

```go
// in lift.go

import "runtime"

func NewAPIError(err error, statusCode int) *APIError {
	return NewAPIErrorWithDepth(err, statusCode, 2) // Default depth
}

func NewAPIErrorWithDepth(err error, statusCode int, depth int) *APIError {
	pc, _, _, _ := runtime.Caller(depth)
	return &APIError{
		err:        err,
		StatusCode: statusCode,
		pc:         pc,
	}
}
```

The existing `lift` function will be updated to use `NewAPIError` when an error is returned from an action, ensuring the PC is captured from the action's frame.

### 3. Update `responder.Error`

The `responder.Error` method in `responder.go` will be updated to check if the error is an `*APIError` containing a captured `pc`. If so, it will resolve the PC to a file and line number and add it as a `slog.Source` attribute to the log record.

```go
// in responder.go

import (
	"log/slog"
	"runtime"
)

func (r *Responder) Error(w http.ResponseWriter, req *http.Request, statusCode int, err error) {
	// ... (existing code)

	var attrs []slog.Attr
	var apiErr *lift.APIError
	if errors.As(err, &apiErr) && apiErr.pc != 0 {
		fs := runtime.CallersFrames([]uintptr{apiErr.pc})
		f, _ := fs.Next()
		if f.File != "" {
			source := &slog.Source{
				File: f.File,
				Line: f.Line,
				// Function: f.Function, // Optional
			}
			attrs = append(attrs, slog.Any("source", source))
		}
	}

	logger.ErrorContext(req.Context(), msg, append(attrs, slog.String("error", err.Error()))...)

	// ... (existing response writing code)
}
```

This change ensures that if an `APIError` is created with position information, that position is logged instead of the location within the `responder.Error` method itself. This will provide much more useful debugging information.

## Future Work: Stack Traces for 500 Errors

To further improve debugging for unexpected (500-level) errors, we can include a full stack trace in the log output.

### Motivation

While the current implementation provides the location where an error was created, a full stack trace offers a much richer context, showing the complete call chain that led to the error. This is invaluable for diagnosing complex, unexpected failures.

### Proposal

1.  **Introduce a Modern Stack Trace Library:** We should adopt an actively maintained library like `gitlab.com/tozd/go/errors`. It serves as a modern, drop-in replacement for the outdated `github.com/pkg/errors`, is fully compatible with Go's standard error wrapping (`%w`), and automatically captures a stack trace when an error is created or wrapped.

2.  **Update Error Handling in `lift`:** Inside the `lift` function, when an unexpected error (one that is not an `*APIError`) is caught, we should wrap it using `errors.Errorf` with the `%w` verb to attach the stack trace.

    ```go
    // in lift.go (inside Lift)
    import "gitlab.com/tozd/go/errors"

    // ...
    data, err := action(r)
    if err != nil {
        // ... (existing error handling for RedirectError and APIError)

        // For unexpected errors, wrap them to capture the stack trace.
        wrappedErr := errors.Errorf("unexpected error in action: %w", err)
        responder.Error(w, r, http.StatusInternalServerError, wrappedErr)
        return
    }
    ```

3.  **Enhance `responder.Error`:** The `responder.Error` method should be updated to log the full stack trace for 500-level errors. The `tozd/go/errors` library, like its predecessor, supports the `%+v` formatting verb for this purpose.

    ```go
    // in responder.go
    import (
        "fmt"
        "log/slog"
    )

    func (r *Responder) Error(w http.ResponseWriter, req *http.Request, statusCode int, err error) {
        // ... (existing logging setup)

        logMsg := "API Error"
        attrs := []slog.Attr{
            slog.Int("status", statusCode),
        }

        if statusCode >= 500 {
            // For server errors, log the full stack trace if available.
            // The `%+v` format verb is supported by libraries like tozd/go/errors.
            attrs = append(attrs, slog.String("error", fmt.Sprintf("%+v", err)))
        } else {
            // For client errors, the simple error message is sufficient.
            attrs = append(attrs, slog.String("error", err.Error()))
        }

        // ... (add source attribute as before)

        logger.LogAttrs(ctx, slog.LevelError, logMsg, attrs...)

        // ... (existing response writing)
    }
    ```

This approach allows us to selectively add detailed stack traces for internal server errors without cluttering the logs for expected client-side errors (4xx), using a modern and maintained library.
