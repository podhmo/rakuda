package rakuda

import (
	"net/http"
	"path"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

type handlerRegistration struct {
	method  string
	pattern string
	handler http.Handler
}

type node struct {
	pattern     string
	middlewares []Middleware
	handlers    []handlerRegistration
	children    []*node
}

// Builder is the configuration object for the router.
// It is used to define routes and middlewares.
// It does not implement http.Handler.
type Builder struct {
	node *node
}

// NewBuilder creates a new Builder instance.
func NewBuilder() *Builder {
	return &Builder{
		node: &node{},
	}
}

func (b *Builder) registerHandler(method string, pattern string, handler http.Handler) {
	b.node.handlers = append(b.node.handlers, handlerRegistration{
		method:  method,
		pattern: pattern,
		handler: handler,
	})
}

// Use adds a middleware to the current builder's node.
func (b *Builder) Use(middleware Middleware) {
	b.node.middlewares = append(b.node.middlewares, middleware)
}

// Get registers a GET handler.
func (b *Builder) Get(pattern string, handler http.Handler) {
	b.registerHandler(http.MethodGet, pattern, handler)
}

// Post registers a POST handler.
func (b *Builder) Post(pattern string, handler http.Handler) {
	b.registerHandler(http.MethodPost, pattern, handler)
}

// Put registers a PUT handler.
func (b *Builder) Put(pattern string, handler http.Handler) {
	b.registerHandler(http.MethodPut, pattern, handler)
}

// Delete registers a DELETE handler.
func (b *Builder) Delete(pattern string, handler http.Handler) {
	b.registerHandler(http.MethodDelete, pattern, handler)
}

// Patch registers a PATCH handler.
func (b *Builder) Patch(pattern string, handler http.Handler) {
	b.registerHandler(http.MethodPatch, pattern, handler)
}

// Route creates a new routing group.
func (b *Builder) Route(pattern string, fn func(b *Builder)) {
	childNode := &node{
		pattern: pattern,
	}
	b.node.children = append(b.node.children, childNode)
	childBuilder := &Builder{node: childNode}
	fn(childBuilder)
}

// Build creates a new http.Handler from the configured routes.
// The returned handler is immutable.
func (b *Builder) Build() http.Handler {
	mux := http.NewServeMux()
	var traverse func(*node, string, []Middleware)
	traverse = func(n *node, prefix string, middlewares []Middleware) {
		// Combine middlewares from parent and current node
		combinedMiddlewares := append([]Middleware{}, middlewares...)
		combinedMiddlewares = append(combinedMiddlewares, n.middlewares...)

		// Register handlers with combined middlewares
		for _, reg := range n.handlers {
			fullPattern := path.Join(prefix, reg.pattern)
			handler := reg.handler
			for i := len(combinedMiddlewares) - 1; i >= 0; i-- {
				handler = combinedMiddlewares[i](handler)
			}
			mux.Handle(reg.method+" "+fullPattern, handler)
		}

		// Traverse children
		for _, child := range n.children {
			newPrefix := path.Join(prefix, child.pattern)
			traverse(child, newPrefix, combinedMiddlewares)
		}
	}

	traverse(b.node, "/", []Middleware{})
	return mux
}
