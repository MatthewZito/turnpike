package turnpike

import (
	"context"
	"net/http"
)

// Router represents a multiplexer that routes HTTP requests.
type Router struct {
	trie                    *trie
	NotFoundHandler         http.Handler
	MethodNotAllowedHandler http.Handler
}

// Route represents a route record to be used by a Router.
type Route struct {
	methods       []string
	path          string
	handler       http.Handler
	middlewares   middlewares
	isFileHandler bool
}

var (
	cachedRoute                    = &Route{}
	DefaultNotFoundHandler         = http.NotFoundHandler
	DefaultMethodNotAllowedHandler = func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		})
	}
)

// NewRouter constructs and returns a pointer to a new Router.
func NewRouter() *Router {
	return &Router{
		trie: newTrie(),
	}
}

// Use adds middlewares to the current Route record.
func (r *Router) Use(mws ...middleware) *Router {
	nm := newMiddlewares(mws)
	cachedRoute.middlewares = nm
	return r
}

// WithMethods appends user-specified HTTP methods to the current Route record.
func (r *Router) WithMethods(methods ...string) *Router {
	cachedRoute.methods = append(cachedRoute.methods, methods...)

	return r
}

// Handler adds a path and handler to the current Route record.
func (r *Router) Handler(path string, handler http.Handler) *Router {
	cachedRoute.path = path
	cachedRoute.handler = handler

	return r
}

// FileHandler registers a route handler as a file server, serving the directory qualified as `path`.
// This method effectively replaces the `Handler()` stage of the route pipeline.
// @todo allow usage of custom NotFoundHandler
func (r *Router) FileHandler(path string, root http.FileSystem) *Router {
	fileServer := http.FileServer(root)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fileServer.ServeHTTP(w, req)
	})

	cachedRoute.isFileHandler = true
	return r.Handler(path, handler).WithMethods(http.MethodGet)
}

// Register registers the current Route record. This method must be invoked to register the Route.
func (r *Router) Register() {
	if len(cachedRoute.methods) == 0 {
		panic("Cannot register a route handler with no specified HTTP methods.")
	}

	if cachedRoute.path == "" || cachedRoute.handler == nil {
		panic("Cannot register a route handler with no specified path or handler.")
	}

	if cachedRoute.isFileHandler && len(cachedRoute.methods) > 1 {
		panic("Cannot register a file route handler with HTTP methods other than GET.")
	}

	r.trie.insert(cachedRoute.methods, cachedRoute.path, cachedRoute.handler, cachedRoute.middlewares)
	cachedRoute = &Route{}
}

// ServeHTTP routes an HTTP request to the appropriate Route record handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path

	result, err := r.trie.search(method, path)
	if err == ErrNotFound {
		if r.NotFoundHandler == nil {
			DefaultNotFoundHandler().ServeHTTP(w, req)
			return
		}
		r.NotFoundHandler.ServeHTTP(w, req)
		return
	}

	if err == ErrMethodNotAllowed {
		if r.MethodNotAllowedHandler == nil {
			DefaultMethodNotAllowedHandler().ServeHTTP(w, req)
			return
		}
		r.MethodNotAllowedHandler.ServeHTTP(w, req)
		return
	}

	handler := result.actions.handler
	// If extant, apply middlewares.
	if result.actions.middlewares != nil {
		handler = result.actions.middlewares.then(result.actions.handler)
	}

	if result.parameters != nil {
		ctx := context.WithValue(req.Context(), parameterKey, result.parameters)
		req = req.WithContext(ctx)
	}

	handler.ServeHTTP(w, req)
}
