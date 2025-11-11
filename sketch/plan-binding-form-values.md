# Plan: Add Form Value Binding

## Goal

Extend the `./binding` package to support binding form values from `application/x-www-form-urlencoded` and `multipart/form-data` requests to Go structs.

## Background

The current `binding` package provides type-safe binding for values from URL query parameters, headers, cookies, and path segments. It lacks support for request bodies, specifically form data, which is a common requirement for handling HTML form submissions.

## Plan

1.  **Introduce a `Form` Source:**
    *   Add a new constant `Form Source = "form"` to the `Source` enum in `binding/binding.go`.

2.  **Update Value Retrieval Logic:**
    *   Modify the `Lookup` and `valuesFromSource` methods in `binding.go` to handle the new `Form` source.
    *   The implementation will call `http.Request.ParseForm()` to parse the request body. This method correctly handles both `application/x-www-form-urlencoded` and `multipart/form-data`.
    *   To ensure a clear distinction from URL query parameters (handled by the `Query` source), the new logic will retrieve values specifically from `r.PostForm`.
    *   The error from `r.ParseForm()` will be implicitly handled. If parsing fails, `r.PostForm` will be empty, causing any `Required` bindings to fail as expected (i.e., "key is required"), which is an acceptable behavior.

3.  **No Public API Changes:**
    *   The existing public functions (`One`, `OnePtr`, `Slice`, `SlicePtr`) will work with the new `Form` source without any changes to their signatures.

4.  **Write Comprehensive Tests:**
    *   In `binding/binding_test.go`, add new test cases to verify the form binding functionality.
    *   Tests will cover:
        *   Both `application/x-www-form-urlencoded` and `multipart/form-data` content types.
        *   Binding of single values (`One`, `OnePtr`).
        *   Binding of multiple values into slices (`Slice`, `SlicePtr`).
        *   Handling of required and optional fields.
        *   Correct behavior for empty or malformed request bodies.
