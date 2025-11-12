package main

import (
	"fmt"
	"log"
	"net/http"

	"flag"
	"os"

	"github.com/podhmo/rakuda"
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

func newRouter() *rakuda.Builder {
	builder := rakuda.NewBuilder()

	// 1. A simple handler that returns a JSON response.
	builder.Get("/", http.HandlerFunc(handleRoot))

	// 2. A handler that uses a path parameter.
	builder.Get("/hello/{name}", http.HandlerFunc(handleHello))

	// 3. A handler that uses rakuda.Lift to simplify returning data and errors.
	builder.Get("/me", rakuda.Lift(responder, actionMe))

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
