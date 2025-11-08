package rakuda

import (
	"net/http"
	"path"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// --- Action definitions ---
type action interface {
	isAction()
}

type middlewareAction struct {
	middleware Middleware
}

func (middlewareAction) isAction() {}

type handlerAction struct {
	method  string
	pattern string
	handler http.Handler
}

func (handlerAction) isAction() {}

// --- Node definition ---
type node struct {
	pattern  string
	actions  []action
	children []*node
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
	b.node.actions = append(b.node.actions, handlerAction{
		method:  method,
		pattern: pattern,
		handler: handler,
	})
}

// Use adds a middleware to the current builder's node.
func (b *Builder) Use(middleware Middleware) {
	b.node.actions = append(b.node.actions, middlewareAction{middleware: middleware})
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

// Group creates a new middleware-only group.
func (b *Builder) Group(fn func(b *Builder)) {
	childNode := &node{}
	b.node.children = append(b.node.children, childNode)
	childBuilder := &Builder{node: childNode}
	fn(childBuilder)
}

// Walk traverses the routing tree and calls the provided function for each registered handler.
// The traversal is done in DFS order.
func (b *Builder) Walk(fn func(method string, pattern string)) {
	var traverse func(*node, string, []Middleware)
	traverse = func(n *node, prefix string, inheritedMiddlewares []Middleware) {
		// Phase 1: Collect middlewares for the current node.
		var nodeMiddlewares []Middleware
		for _, a := range n.actions {
			if ma, ok := a.(middlewareAction); ok {
				nodeMiddlewares = append(nodeMiddlewares, ma.middleware)
			}
		}

		// Combine inherited middlewares with the current node's middlewares.
		combinedMiddlewares := append([]Middleware{}, inheritedMiddlewares...)
		combinedMiddlewares = append(combinedMiddlewares, nodeMiddlewares...)

		// Phase 2: call fn for each handler.
		for _, a := range n.actions {
			if ha, ok := a.(handlerAction); ok {
				fullPattern := path.Join(prefix, ha.pattern)
				fn(ha.method, fullPattern)
			}
		}

		// Phase 3: Traverse children.
		for _, child := range n.children {
			newPrefix := path.Join(prefix, child.pattern)
			traverse(child, newPrefix, combinedMiddlewares)
		}
	}

	traverse(b.node, "/", []Middleware{})
}

// Build creates a new http.Handler from the configured routes.
// The returned handler is immutable.
func (b *Builder) Build() http.Handler {
	mux := http.NewServeMux()
	var traverse func(*node, string, []Middleware)
	traverse = func(n *node, prefix string, inheritedMiddlewares []Middleware) {
		// Phase 1: Collect middlewares for the current node.
		var nodeMiddlewares []Middleware
		for _, a := range n.actions {
			if ma, ok := a.(middlewareAction); ok {
				nodeMiddlewares = append(nodeMiddlewares, ma.middleware)
			}
		}

		// Combine inherited middlewares with the current node's middlewares.
		combinedMiddlewares := append([]Middleware{}, inheritedMiddlewares...)
		combinedMiddlewares = append(combinedMiddlewares, nodeMiddlewares...)

		// Phase 2: Register handlers with the combined middleware chain.
		for _, a := range n.actions {
			if ha, ok := a.(handlerAction); ok {
				fullPattern := path.Join(prefix, ha.pattern)
				handler := ha.handler
				for i := len(combinedMiddlewares) - 1; i >= 0; i-- {
					handler = combinedMiddlewares[i](handler)
				}
				mux.Handle(ha.method+" "+fullPattern, handler)
			}
		}

		// Phase 3: Traverse children.
		for _, child := range n.children {
			newPrefix := path.Join(prefix, child.pattern)
			traverse(child, newPrefix, combinedMiddlewares)
		}
	}

	traverse(b.node, "/", []Middleware{})
	return mux
}
