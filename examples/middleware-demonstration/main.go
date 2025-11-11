package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/podhmo/rakuda"
	"github.com/podhmo/rakuda/rakudamiddleware"
)

func newRouter() http.Handler {
	builder := rakuda.NewBuilder()

	// Use the recovery middleware globally
	builder.Use(rakudamiddleware.Recovery)

	builder.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world!")
	}))

	// This handler will panic, but the Recovery middleware will catch it.
	builder.Get("/panic", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	handler, err := builder.Build()
	if err != nil {
		panic(err) // In a real app, you'd handle this more gracefully.
	}
	return handler
}

func main() {
	handler := newRouter()
	port := 8080
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	logger.InfoContext(context.Background(), "server starting", "port", port)
	logger.InfoContext(context.Background(), "Try accessing http://localhost:8080/ or http://localhost:8080/panic")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		logger.ErrorContext(context.Background(), "server failed", "error", err)
		os.Exit(1)
	}
}
