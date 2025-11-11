# Package binding

The `binding` package provides a type-safe, reflect-free, and expression-oriented way to bind data from HTTP requests to Go structs.

## Usage

Here is a simple example of how to use the `binding` package to extract data from a request query, header, and path parameter.

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/podhmo/rakuda/binding"
)

// Define a parser for converting a string to an int.
var parseInt = func(s string) (int, error) {
	return strconv.Atoi(s)
}

// Define a simple string parser.
var parseString = func(s string) (string, error) {
	return s, nil
}

type MyParams struct {
	ID      int
	Token   string
	SortKey *string // Optional
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	// In a real application, the pathValue function would be provided by your router.
	// For this example, we'll simulate it.
	pathValue := func(key string) string {
		if key == "id" {
			// Simulate extracting "123" from a path like "/users/123"
			return "123"
		}
		return ""
	}

	// Create a new binder for the current request.
	b := binding.New(r, pathValue)

	var params MyParams
	if err := errors.Join(
		binding.One(b, &params.ID, binding.Path, "id", parseInt, binding.Required),
		binding.One(b, &params.Token, binding.Header, "X-Auth-Token", parseString, binding.Required),
		binding.OnePtr(b, &params.SortKey, binding.Query, "sort", parseString, binding.Optional),
	); err != nil {
		http.Error(w, fmt.Sprintf("Bad Request: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Successfully bound parameters:\n")
	fmt.Fprintf(w, "ID: %d\n", params.ID)
	fmt.Fprintf(w, "Token: %s\n", params.Token)
	if params.SortKey != nil {
		fmt.Fprintf(w, "Sort Key: %s\n", *params.SortKey)
	} else {
		fmt.Fprintf(w, "Sort Key: not provided\n")
	}
}

func main() {
	http.HandleFunc("/users/123", myHandler)
	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
```

### How to Run the Example

1.  Save the code above as `main.go`.
2.  Run `go mod init example` and `go mod tidy`.
3.  Run `go run main.go`.
4.  Send a request to the server:

    ```sh
    # Request with all parameters
    curl -H "X-Auth-Token: mysecret" "http://localhost:8080/users/123?sort=name"

    # Request without the optional sort parameter
    curl -H "X-Auth-Token: mysecret" "http://localhost:8080/users/123"

    # Request missing a required parameter (will result in a 400 Bad Request)
    curl "http://localhost:8080/users/123?sort=name"
    ```
