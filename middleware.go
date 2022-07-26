package turnpike

import "net/http"

// middleware represents a singular instance of a Route handler middleware.
type middleware func(http.Handler) http.Handler

// middlewares represents a slice of middleware.
type middlewares []middleware

// newMiddlewares creates and returns middlewares.
func newMiddlewares(mws middlewares) middlewares {
	return append([]middleware(nil), mws...)
}

// Then executes middlewares.
func (mws middlewares) then(h http.Handler) http.Handler {
	for i := range mws {
		h = mws[len(mws)-1-i](h)
	}

	return h
}
