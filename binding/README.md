# Package binding

The `binding` package provides a type-safe, reflect-free, and expression-oriented way to bind data from HTTP requests to Go structs. It is designed to produce structured, user-friendly error responses out of the box when integrated with the `rakuda` Lift and Responder components.

## Features

- **Type-Safe**: Uses generics to ensure type safety at compile time.
- **Reflect-Free**: Avoids reflection for better performance and clearer code.
- **Structured Errors**: Automatically generates detailed JSON error responses for validation failures.
- **Extensible**: Supports custom parsers for any data type.

## Usage

Here is a simple example of how to use the `binding` package within a `rakuda` application. The `binding.Join` function aggregates any validation errors, and `rakuda.Lift` handles the error by sending a structured JSON response.

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/podhmo/rakuda"
	"github.com/podhmo/rakuda/binding"
)

var responder = rakuda.NewResponder()

// Define a parser for converting a string to an int.
var parseInt = func(s string) (int, error) {
	return strconv.Atoi(s)
}

// Define a simple string parser.
var parseString = func(s string) (string, error) {
	return s, nil
}

// GistParams represents the parameters for the gist action.
type GistParams struct {
	ID    int     `json:"id"`
	Token string  `json:"-"` // Not included in JSON response
	Sort  *string `json:"sort,omitempty"`
}

// actionGist is a handler that demonstrates data binding.
func actionGist(r *http.Request) (GistParams, error) {
	var params GistParams
	b := binding.New(r, r.PathValue)

	// Use binding.Join to aggregate all validation errors.
	if err := binding.Join(
		binding.One(b, &params.ID, binding.Path, "id", parseInt, binding.Required),
		binding.One(b, &params.Token, binding.Header, "X-Auth-Token", parseString, binding.Required),
		binding.OnePtr(b, &params.Sort, binding.Query, "sort", parseString, binding.Optional),
	); err != nil {
		// Lift will automatically handle the *binding.ValidationErrors
		// and the responder will format it into a detailed JSON response.
		return params, err
	}

	return params, nil
}

func main() {
	builder := rakuda.NewBuilder()
	builder.Get("/gists/{id}", rakuda.Lift(responder, actionGist))

	handler, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
```

### How to Run the Example

1.  Save the code above as `main.go`.
2.  Run `go mod init example` and `go mod tidy`.
3.  Run `go run main.go`.
4.  Send requests to the server:

    ```sh
    # 1. Request with all parameters (Success)
    curl -i -H "X-Auth-Token: mysecret" "http://localhost:8080/gists/123?sort=name"
    # HTTP/1.1 200 OK
    # {"id":123,"sort":"name"}

    # 2. Request without the optional sort parameter (Success)
    curl -i -H "X-Auth-Token: mysecret" "http://localhost:8080/gists/456"
    # HTTP/1.1 200 OK
    # {"id":456}

    # 3. Request with multiple validation errors (Failure)
    # - Path parameter 'id' is not a valid integer.
    # - Required 'X-Auth-Token' header is missing.
    curl -i "http://localhost:8080/gists/invalid-id?sort=name"
    # HTTP/1.1 400 Bad Request
    # {
    #   "errors": [
    #     {
    #       "source": "path",
    #       "key": "id",
    #       "value": "invalid-id",
    #       "message": "strconv.Atoi: parsing \"invalid-id\": invalid syntax"
    #     },
    #     {
    #       "source": "header",
    #       "key": "X-Auth-Token",
    #       "value": null,
    #       "message": "required parameter is missing"
    #     }
    #   ]
    # }
    ```
