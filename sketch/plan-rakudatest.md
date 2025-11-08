# Plan: Test Helper `rakudatest`

This document outlines the plan to create a new test helper package `rakudatest` for the `rakuda` project. The design is inspired by `quickapitest` and focuses on a simple, functional API for testing `http.Handler` instances.

The primary goal is to provide a testing utility that ensures response bodies can be inspected even when assertions on status codes fail, preventing loss of debugging information.

## Core Components

### 1. `rakudatest` Package

- Create a new directory and package named `rakudatest`.
- The package will contain all the testing helper functions.
- It will have its own test file, `rakudatest_test.go`, to ensure the helpers work as expected.

### 2. Core `Do` Function

- **`Do(t *testing.T, h http.Handler, req *http.Request) (*http.Response, []byte)`**
    - This will be the central function.
    - It executes the request against the handler using `httptest.NewRecorder`.
    - It reads the entire response body into a `[]byte` slice.
    - It closes the response body to prevent resource leaks.
    - It returns both the original `*http.Response` (with a now-closed body) and the `[]byte` slice of the body. This allows assertions on headers/status from the response object, and assertions on the body from the byte slice.

### 3. Request Builder Helpers (Functional Options)

- **`NewRequest(method, path string, options ...RequestOption) *http.Request`**
    - Creates a new `*http.Request` for testing.
- **`RequestOption` type: `func(*http.Request) error`**
    - A functional option type to modify the request.
- **Implement common options:**
    - `WithBody(io.Reader)`: Sets the request body.
    - `WithJSONBody(any)`: Marshals the given value to JSON and sets it as the body, also setting the `Content-Type` header to `application/json`.
    - `WithHeader(key, value string)`: Sets a request header.
    - `WithPathValue(key, value string)`: Adds a path value to the request context (requires Go 1.22+).

### 4. Assertion Functions

These functions will be standalone and use the standard `testing` package.

- **`AssertStatusCode(t *testing.T, res *http.Response, body []byte, want int)`**
    - Checks if `res.StatusCode` matches `want`.
    - If it fails, it will log the response body for easier debugging. This is a key requirement.
- **`AssertHeader(t *testing.T, res *http.Response, key, want string)`**
    - Checks for the presence and value of a response header.
- **`AssertJSON[T any](t *testing.T, body []byte, want T, opts ...cmp.Option)`**
    - Decodes the JSON `body` into a new variable of type `T`.
    - Compares the decoded value with `want` using `cmp.DeepEqual` or `cmp.Equal`.
    - The function will be generic.

## Implementation Steps (TODO)

1.  **Create `rakudatest/rakudatest.go`**:
    -   Create the new directory and file.
    -   Define the main `Do` function.
    -   Implement the request builder `NewRequest` and `RequestOption` helpers (`WithJSONBody`, `WithHeader`, etc.).
2.  **Implement Assertion Functions**:
    -   Add `AssertStatusCode`, `AssertHeader`, and `AssertJSON` to `rakudatest/rakudatest.go`. Make sure `AssertStatusCode` logs the body on failure.
3.  **Create Tests for the Helper (`rakudatest/rakudatest_test.go`)**:
    -   Write tests to verify that `Do`, the request options, and the assertion helpers all work correctly.
    -   Test a failing case to ensure the response body is logged.
4.  **Refactor an Existing Test**:
    -   Choose one existing test file (e.g., `builder_test.go` or `lift_test.go`).
    -   Update a test case to use the new `rakudatest` helpers. This will demonstrate its usage and prove its utility.
5.  **Update `TODO.md`**:
    -   Transcribe these steps into `TODO.md` under a new `### rakudatest Test Helper` section.
