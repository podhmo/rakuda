package main

import (
	"fmt"
	"log"
	"net/http"

	"flag"
	"os"

	"github.com/podhmo/rakuda"
)

func newRouter() *rakuda.Builder {
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
