package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"flag"
	"os"

	"github.com/podhmo/rakuda"
	"github.com/podhmo/rakuda/binding"
)

var responder = rakuda.NewResponder()

// handleRoot is a simple handler that returns a JSON response.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	responder.JSON(w, r, http.StatusOK, map[string]string{"message": "hello world"})
}

// handleHello is a handler that uses a path parameter.
func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	responder.JSON(w, r, http.StatusOK, map[string]string{"message": fmt.Sprintf("hello %s", name)})
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// actionMe is an action that returns a user object.
func actionMe(r *http.Request) (User, error) {
	// In a real application, you would fetch the user from a database.
	// This is just a simple example.
	user := User{ID: 1, Name: "foo"}
	return user, nil
}

// actionGist is an action that demonstrates data binding and structured error responses.
type Gist struct {
	ID    int     `json:"id"`
	Token string  `json:"-"`
	Sort  *string `json:"sort,omitempty"`
}

func actionGist(r *http.Request) (Gist, error) {
	var params Gist
	b := binding.New(r, r.PathValue)

	if err := binding.Join(
		binding.One(b, &params.ID, binding.Path, "id", strconv.Atoi, binding.Required),
		binding.One(b, &params.Token, binding.Header, "X-Auth-Token", func(s string) (string, error) { return s, nil }, binding.Required),
		binding.OnePtr(b, &params.Sort, binding.Query, "sort", func(s string) (string, error) { return s, nil }, binding.Optional),
	); err != nil {
		// Lift will automatically handle the *binding.ValidationErrors
		// and the responder will format it into a detailed JSON response.
		return params, err
	}

	if params.Token != "secret" {
		return params, rakuda.NewAPIError(http.StatusUnauthorized, errors.New("invalid token"))
	}

	return params, nil
}

func newRouter() *rakuda.Builder {
	builder := rakuda.NewBuilder()

	// 1. A simple handler that returns a JSON response.
	builder.Get("/", http.HandlerFunc(handleRoot))

	// 2. A handler that uses a path parameter.
	builder.Get("/hello/{name}", http.HandlerFunc(handleHello))

	// 3. A handler that uses rakuda.Lift to simplify returning data and errors.
	builder.Get("/me", rakuda.Lift(responder, actionMe))

	// 4. A handler demonstrating data binding with structured error responses.
	builder.Get("/gists/{id}", rakuda.Lift(responder, actionGist))

	return builder
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("!%+v", err)
	}
}

func run() error {
	var (
		proutes = flag.Bool("proutes", false, "print routes")
		port    = flag.Int("port", 8080, "port")
	)
	flag.Parse()

	builder := newRouter()
	if *proutes {
		rakuda.PrintRoutes(os.Stdout, builder)
		return nil
	}

	handler, err := builder.Build()
	if err != nil {
		return err
	}
	log.Printf("listening on :%d", *port)
	return http.ListenAndServe(fmt.Sprintf(":%d", *port), handler)
}
