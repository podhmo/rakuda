# Plan: Responder methods should have status code

- [x] Responder's methods (not sse) should take status code
- [x] `Responder.Redirect()` should redirect as it is
- [x] remove `WithStatusCode`, `StatusCodeFromContext`
- [x] update README.md
- [x] update examples
- [x] fix tests

## Details

### Responder's methods (not sse) should take status code

e.g.

- `func (r *Responder) JSON(w http.ResponseWriter, req *http.Request, statusCode int, ob any)`
- `func (r *Responder) Error(w http.ResponseWriter, req *http.Request, statusCode int, err error)`

### `Responder.Redirect()` should redirect as it is

`Responder.Redirect()` is just `http.Redirect`.

In `Lift`, `RedirectError` is used for redirection.

```go
func(w http.ResponseWriter, r *http.Request) (any, error) {
    // ...
    return nil, &RedirectError{URL: "/login", StatusCode: http.StatusFound}
}
```

This also means that the function passed to `Lift` does not need to use the Responder's methods.
