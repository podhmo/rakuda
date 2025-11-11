package rakudamiddleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/podhmo/rakuda"
)

// CORSConfig holds the configuration for CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to access the resource.
	// Use "*" to allow any origin. Default is "*".
	AllowedOrigins []string
	// AllowedMethods is a list of methods the client is allowed to use.
	// Default is GET, POST, PUT, DELETE, PATCH, OPTIONS.
	AllowedMethods []string
	// AllowedHeaders is a list of headers the client is allowed to use.
	// Default is Accept, Content-Type, Authorization.
	AllowedHeaders []string
	// AllowCredentials indicates whether the request can include user credentials.
	// Default is false.
	AllowCredentials bool
	// MaxAge indicates how long the results of a preflight request can be cached.
	// Default is 3600 seconds (1 hour).
	MaxAge int
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing (CORS).
// If config is nil, it uses default permissive settings.
func CORS(config *CORSConfig) rakuda.Middleware {
	if config == nil {
		config = &CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
			MaxAge:         3600,
		}
	}

	// Set defaults if not specified
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Accept", "Content-Type", "Authorization"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 3600
	}

	allowedMethods := strings.Join(config.AllowedMethods, ", ")
	allowedHeaders := strings.Join(config.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" {
				isAllowed := false
				for _, allowedOrigin := range config.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						isAllowed = true
						break
					}
				}

				if isAllowed {
					if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					} else {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Vary", "Origin")
					}

					if config.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}
				}
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
