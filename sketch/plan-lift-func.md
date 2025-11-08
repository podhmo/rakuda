# Plan: Implement `lift` function

## Goal

Create a `lift` function that converts a function with the signature `func[O any](*Request) (O, error)` into a standard `http.Handler`.

This will simplify handler logic by allowing developers to return data and an error directly, with the `lift` function handling the boilerplate of JSON encoding, error handling, and status code management.

## Detailed Steps

1.  **Define the `lift` function signature:**
    -   It will be a generic function: `lift[O any](responder *Responder, action func(*http.Request) (O, error)) http.Handler`.
    -   It will take a `*Responder` instance to handle JSON encoding and logging.
    -   It will take an `action` function, which is the user-provided handler logic.
    -   It will return an `http.Handler` that can be used with `rakuda.Builder` or standard `net/http`.

2.  **Implement the `lift` function logic:**
    -   The returned `http.Handler` will be a `http.HandlerFunc`.
    -   Inside the handler function:
        -   Call the `action` function with the `*http.Request`.
        -   Check the returned `error`.
        -   **If `error` is `nil`**:
            -   Call `responder.JSON()` to encode the returned data object (`O`) with a status of `http.StatusOK`.
        -   **If `error` is not `nil`**:
            -   Inspect the error to see if it carries a status code (e.g., by using a custom error type or `errors.As`).
            -   If a status code is found, use it.
            -   If no status code is found, default to `http.StatusInternalServerError`.
            -   Create a structured error response (e.g., `map[string]string{"error": err.Error()}`).
            -   Use `WithStatusCode` to set the appropriate status on the request context.
            -   Call `responder.JSON()` to send the error response.

3.  **Define a custom error type for status codes:**
    -   Create a new type, e.g., `HTTPError`, that embeds an `error` and includes a `StatusCode` field.
    -   This will allow `lift` to easily extract the status code from errors returned by the `action` function.

4.  **Add comprehensive tests for the `lift` function:**
    -   Create `lift_test.go`.
    -   **Test Case 1: Success**
        -   `action` returns a valid data object and `nil` error.
        -   Verify the response is a 200 OK with the correct JSON body.
    -   **Test Case 2: Error with Status Code**
        -   `action` returns `nil` data and an `HTTPError` with a specific status code (e.g., 400 Bad Request).
        -   Verify the response has the correct status code and a JSON error body.
    -   **Test Case 3: Error without Status Code**
        -   `action` returns `nil` data and a standard `error` (e.g., `errors.New("something went wrong")`).
        -   Verify the response is a 500 Internal Server Error with a JSON error body.
    -   **Test Case 4: Nil data, nil error**
        -    `action` returns `nil`, `nil`.
        -   Verify the response is a 200 OK with a `null` JSON body.

5.  **Update `TODO.md`:**
    -   Add a new task for the `lift` function implementation under a suitable section.
    -   Mark the sub-tasks as they are completed.
    -   Finally, mark the main task as complete.
