package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/podhmo/rakuda"
	"github.com/podhmo/rakuda/binding"
	"github.com/podhmo/rakuda/rakudamiddleware"
)

//go:embed static/*
var staticFiles embed.FS

// Parser for string values
var parseString binding.Parser[string] = func(s string) (string, error) {
	return s, nil
}

// Structs for binding
type UserIDParams struct {
	ID string
}

type AuthHeader struct {
	Authorization string
}

func newRouter() *rakuda.Builder {
	builder := rakuda.NewBuilder()
	responder := rakuda.NewResponder()

	// Global middleware: Recovery for all routes
	builder.Use(rakudamiddleware.Recovery)

	// Serve static files from embedded filesystem
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to create sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticFS))
	builder.Get("/static/{path...}", http.StripPrefix("/static/", fileServer))

	// Serve index.html at root
	builder.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}))

	// API routes
	builder.Route("/api", func(api *rakuda.Builder) {
		// Add request logging middleware for API routes
		api.Use(loggingMiddleware())

		// Public routes (no additional middleware)
		api.Route("/public", func(public *rakuda.Builder) {
			public.Get("/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				responder.JSON(w, r, map[string]any{
					"name":        "Rakuda SPA API",
					"version":     "1.0.0",
					"description": "Example SPA with go:embed and CORS support",
					"timestamp":   time.Now().Format(time.RFC3339),
				})
			}))
		})

		// User routes (with auth middleware)
		api.Route("/users", func(users *rakuda.Builder) {
			users.Use(authMiddleware())

			users.Get("/current", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user := r.Context().Value("user")
				responder.JSON(w, r, map[string]any{
					"user":    user,
					"message": "Successfully retrieved current user",
				})
			}))

			users.Get("/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Use binding to extract path parameter
				b := binding.New(r, r.PathValue)
				var params UserIDParams
				if err := binding.One(b, &params.ID, binding.Path, "id", parseString, binding.Required); err != nil {
					ctx := rakuda.NewContextWithStatusCode(r.Context(), http.StatusBadRequest)
					r = r.WithContext(ctx)
					responder.JSON(w, r, map[string]string{
						"error": err.Error(),
					})
					return
				}

				responder.JSON(w, r, map[string]any{
					"id":       params.ID,
					"name":     fmt.Sprintf("User %s", params.ID),
					"email":    fmt.Sprintf("user%s@example.com", params.ID),
					"joined":   "2024-01-15",
					"verified": true,
				})
			}))

			users.Post("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := rakuda.NewContextWithStatusCode(r.Context(), http.StatusCreated)
				r = r.WithContext(ctx)
				responder.JSON(w, r, map[string]any{
					"message": "User created successfully",
					"id":      "new-user-id",
				})
			}))
		})

		// Admin routes (with auth + admin middleware)
		api.Route("/admin", func(admin *rakuda.Builder) {
			admin.Use(authMiddleware())
			admin.Use(adminOnlyMiddleware())

			admin.Get("/stats", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				responder.JSON(w, r, map[string]any{
					"total_users":    1337,
					"active_users":   892,
					"total_requests": 42000,
					"uptime":         "7d 12h 34m",
					"version":        "1.0.0",
				})
			}))
		})
	})

	return builder
}

// loggingMiddleware logs incoming requests
func loggingMiddleware() rakuda.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			// Add logger to request context
			ctx := rakuda.NewContextWithLogger(r.Context(), logger)
			r = r.WithContext(ctx)

			logger.InfoContext(r.Context(), "request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			next.ServeHTTP(w, r)

			logger.InfoContext(r.Context(), "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", time.Since(start).String(),
			)
		})
	}
}

// authMiddleware simulates authentication
func authMiddleware() rakuda.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use binding to extract Authorization header
			b := binding.New(r, r.PathValue)
			var auth AuthHeader
			err := binding.One(b, &auth.Authorization, binding.Header, "Authorization", parseString, binding.Optional)
			if err != nil {
				// If there's an error parsing (unlikely for string), just continue
				next.ServeHTTP(w, r)
				return
			}

			// Simple token validation (for demo purposes)
			if auth.Authorization != "" && len(auth.Authorization) > 7 {
				// Extract token and simulate user lookup
				token := auth.Authorization[7:] // Remove "Bearer " prefix
				ctx := context.WithValue(r.Context(), "user", map[string]any{
					"id":    "user-123",
					"name":  "Demo User",
					"email": "demo@example.com",
					"token": token,
				})
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// adminOnlyMiddleware checks if the user has admin privileges
func adminOnlyMiddleware() rakuda.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value("user")
			if user == nil {
				ctx := rakuda.NewContextWithStatusCode(r.Context(), http.StatusUnauthorized)
				r = r.WithContext(ctx)
				responder := rakuda.NewResponder()
				responder.JSON(w, r, map[string]string{
					"error": "Authentication required",
				})
				return
			}

			// Simulate admin check (in real app, check user role from database)
			userMap, ok := user.(map[string]any)
			if !ok {
				ctx := rakuda.NewContextWithStatusCode(r.Context(), http.StatusForbidden)
				r = r.WithContext(ctx)
				responder := rakuda.NewResponder()
				responder.JSON(w, r, map[string]string{
					"error": "Insufficient permissions",
				})
				return
			}

			// For demo: tokens containing "admin" are considered admin tokens
			token, _ := userMap["token"].(string)
			if token == "" || len(token) < 5 || token[:5] != "admin" {
				ctx := rakuda.NewContextWithStatusCode(r.Context(), http.StatusForbidden)
				r = r.WithContext(ctx)
				responder := rakuda.NewResponder()
				responder.JSON(w, r, map[string]string{
					"error": "Admin access required",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
		return fmt.Errorf("failed to build router: %w", err)
	}

	// Wrap the entire handler with CORS to catch all OPTIONS requests
	// This ensures preflight requests are handled even for routes not explicitly registered
	corsHandler := rakudamiddleware.CORS(&rakudamiddleware.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           3600,
	})(handler)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.InfoContext(context.Background(), "server starting",
		"port", *port,
		"url", fmt.Sprintf("http://localhost:%d", *port),
	)
	logger.InfoContext(context.Background(), "Open your browser and visit the URL above")

	return http.ListenAndServe(fmt.Sprintf(":%d", *port), corsHandler)
}
