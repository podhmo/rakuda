# Plan: Implement `Lift` function

> [!NOTE]
> This feature has been implemented.

## Goal

Create a `Lift` function that converts a function with the signature `func[O any](*Request) (O, error)` into a standard `http.Handler`.

This will simplify handler logic by allowing developers to return data and an error directly, with the `Lift` function handling the boilerplate of JSON encoding, error handling, and status code management.

## Detailed Steps

1.  **Define the `Lift` function signature:**
    -   It will be a public, generic function: `Lift[O any](responder *Responder, action func(*http.Request) (O, error)) http.Handler`.
    -   It will take a `*Responder` instance to handle JSON encoding and logging.
    -   It will take an `action` function, which is the user-provided handler logic.
    -   It will return an `http.Handler` that can be used with `rakuda.Builder` or standard `net/http`.

2.  **Implement the `Lift` function logic:**
    -   The returned `http.Handler` will be a `http.HandlerFunc`.
    -   Inside the handler function:
        -   Call the `action` function with the `*http.Request`.
        -   Check the returned `error`.
        -   **If `error` is `nil`**:
            -   Call `responder.JSON()` to encode the returned data object (`O`) with a status of `http.StatusOK`.
        -   **If `error` is not `nil`**:
            -   Inspect the error to see if it implements an interface with a `StatusCode() int` method.
            -   If a status code is found, use it and return the error message in the JSON response.
            -   If no status code is found (i.e., it's an internal error), default to `http.StatusInternalServerError`.
            -   For internal errors, log the detailed error using the logger from the request context (or the responder's default logger) and return a generic, fixed error message (e.g., "Internal Server Error") in the JSON response to avoid leaking implementation details.

3.  **Define a custom error type for status codes:**
    -   Create a new public type, `APIError`, that wraps an `error` and includes a status code.
    -   It will implement the `error` interface and provide a `StatusCode() int` method.
    -   This allows `Lift` to easily extract the status code from errors returned by the `action` function.

4.  **Add comprehensive tests for the `Lift` function:**
    -   Create `lift_test.go`.
    -   **Test Case 1: Success**
        -   `action` returns a valid data object and `nil` error.
        -   Verify the response is a 200 OK with the correct JSON body.
    -   **Test Case 2: Error with Status Code**
        -   `action` returns `nil` data and an `APIError` with a specific status code (e.g., 400 Bad Request).
        -   Verify the response has the correct status code and a JSON error body with the original error message.
    -   **Test Case 3: Error without Status Code**
        -   `action` returns `nil` data and a standard `error` (e.g., `errors.New("something went wrong")`).
        -   Verify the response is a 500 Internal Server Error with a generic JSON error body (e.g., `{"error":"Internal Server Error"}`).
    -   **Test Case 4: Nil data, nil error**
        -    `action` returns `nil`, `nil`.
        -   Verify the response is a 200 OK with a `null` or zero-value JSON body, depending on the return type.

5.  **Update documentation:**
    -   Update `TODO.md` to reflect the completion of the task.
    -   Update `README.md` to include a section explaining how to use the `Lift` function with a clear code example.
