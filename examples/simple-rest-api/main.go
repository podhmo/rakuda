package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/podhmo/rakuda"
)

func newRouter() http.Handler {
	builder := rakuda.NewBuilder()
	responder := rakuda.NewResponder()

	// 1. A simple handler that returns a JSON response.
	builder.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responder.JSON(w, r, map[string]string{"message": "hello world"})
	}))

	// 2. A handler that uses a path parameter.
	builder.Get("/hello/{name}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		responder.JSON(w, r, map[string]string{"message": fmt.Sprintf("hello %s", name)})
	}))

	return builder.Build()
}

func main() {
	handler := newRouter()
	port := 8080
	log.Printf("listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		log.Fatalf("!%+v", err)
	}
}
