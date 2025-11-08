package rakuda

import "net/http"

type handlerRegistration struct {
	method  string
	pattern string
	handler http.Handler
}

type node struct {
	pattern     string
	middlewares []func(http.Handler) http.Handler
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

// Build creates a new http.Handler from the configured routes.
// The returned handler is immutable.
func (b *Builder) Build() http.Handler {
	mux := http.NewServeMux()
	for _, registration := range b.node.handlers {
		mux.Handle(registration.method+" "+registration.pattern, registration.handler)
	}
	return mux
}
