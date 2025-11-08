> [!NOTE]
> This feature has been implemented.

# Plan: Test Helper `rakudatest`

This document outlines the plan to create a new test helper package `rakudatest` for the `rakuda` project. The design is inspired by `quickapitest` and focuses on a single, convenient function for testing `http.Handler` instances.

The primary goal is to provide a testing utility that integrates request execution, status code validation, and response decoding into a single call, while ensuring that the full response body is logged if the status code check fails.

## Core Components

### 1. `rakudatest` Package

- A new package named `rakudatest` containing the test helper.

### 2. Core `Do` Function

- **`Do[T any](t *testing.T, h http.Handler, req *http.Request, wantStatusCode int, assertions ...ResponseAssertion) T`**
    - This is the central generic function of the package.
    - It executes an HTTP request against a handler.
    - It validates that the response status code matches `wantStatusCode`. If it does not, it logs the entire response body and fails the test immediately via `t.Fatalf`.
    - It executes one or more optional `ResponseAssertion` functions to allow for additional checks (e.g., on response headers) before the response body is decoded.
    - It decodes the JSON response body into the specified generic type `T` and returns it.
    - It correctly handles `204 No Content` responses by returning a zero value for `T` without attempting to decode the (empty) body.

### 3. Response Assertion Helper

- **`ResponseAssertion func(t *testing.T, res *http.Response, body []byte)`**
    - A function type that allows for creating custom validation logic for the response.
    - Since the `Do` function consumes the response body, this provides the necessary hook to inspect the `*http.Response` and the raw `body []byte` before the final JSON decoding step.

## Rationale for Design Changes

Through iterative feedback, the design was simplified to remove all request-side helpers (`RequestOption`, `With...` functions). The user is expected to construct the `*http.Request` using the standard library's `httptest.NewRequest` and modify it directly before passing it to the `Do` function. This avoids thin wrappers and keeps the helper's API surface minimal and focused. The addition of `ResponseAssertion` provides the necessary flexibility for response validation that was lost by consuming the response inside a single helper function.

## Implementation Steps

1.  **Create `rakudatest/rakudatest.go`**:
    -   Implement the generic `Do` function with integrated status checking, `ResponseAssertion` execution, and JSON decoding.
2.  **Create Tests for the Helper (`rakudatest/rakudatest_test.go`)**:
    -   Write tests to verify the complete functionality of the `Do` function, including success cases, status code mismatch failures, and the correct execution of `ResponseAssertion` functions.
3.  **Refactor `lift_test.go`**:
    -   Update `lift_test.go` to use the new `rakudatest.Do` function, demonstrating its ability to simplify test code.
4.  **Update `TODO.md`**:
    -   Transcribe these steps into `TODO.md` and mark them as complete upon finishing the implementation.
    -   Move the feature to the "Implemented" section.
