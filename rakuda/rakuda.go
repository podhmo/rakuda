package rakuda

import "net/http"

type handlerRegistration struct {
	method  string
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

func (b *Builder) registerHandler(method string, handler http.Handler) {
	b.node.handlers = append(b.node.handlers, handlerRegistration{
		method:  method,
		handler: handler,
	})
}

// Get registers a GET handler.
func (b *Builder) Get(handler http.Handler) {
	b.registerHandler(http.MethodGet, handler)
}

// Post registers a POST handler.
func (b *Builder) Post(handler http.Handler) {
	b.registerHandler(http.MethodPost, handler)
}

// Put registers a PUT handler.
func (b *Builder) Put(handler http.Handler) {
	b.registerHandler(http.MethodPut, handler)
}

// Delete registers a DELETE handler.
func (b *Builder) Delete(handler http.Handler) {
	b.registerHandler(http.MethodDelete, handler)
}

// Patch registers a PATCH handler.
func (b *Builder) Patch(handler http.Handler) {
	b.registerHandler(http.MethodPatch, handler)
}

// Build creates a new http.Handler from the configured routes.
// The returned handler is immutable.
func (b *Builder) Build() http.Handler {
	// For now, this is a stub.
	return http.NewServeMux()
}
