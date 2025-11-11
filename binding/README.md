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
	// Create a new binder for the current request, passing r.PathValue directly.
	b := binding.New(r, r.PathValue)

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
	mux := http.NewServeMux()
	mux.HandleFunc("/users/{id}", myHandler)

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
```

### How to Run the Example

1.  Save the code above as `main.go`.
2.  Run `go mod init example` and `go mod tidy`. You may need to run `go get github.com/podhmo/rakuda/binding` as well.
3.  Run `go run main.go`.
4.  Send a request to the server:

    ```sh
    # Request with all parameters
    curl -H "X-Auth-Token: mysecret" "http://localhost:8080/users/123?sort=name"

    # Request without the optional sort parameter
    curl -H "X-Auth-Token: mysecret" "http://localhost:8080/users/456"

    # Request missing a required parameter (will result in a 400 Bad Request)
    curl "http://localhost:8080/users/789?sort=name"
    ```
