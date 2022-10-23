# Turnpike

A trie-based HTTP multiplexer for Go with support for regex matching and middleware.

## Usage

```go
const (
	PathRoot              = "/"
	PathDelimiter         = PathRoot
	ParameterDelimiter    = ":"
	PatternDelimiterStart = "["
	PatternDelimiterEnd   = "]"
	PatternWildcard       = "(.+)"
)
```

```go
var (
	ErrNotFound         = errors.New("no matching route record found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)
```

```go
var (
	DefaultNotFoundHandler         = http.NotFoundHandler
	DefaultMethodNotAllowedHandler = func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		})
	}
)
```

#### func  GetParam

```go
func GetParam(ctx context.Context, key string) string
```
GetParam retrieves from context a value corresponding to a given key.

#### type Route

```go
type Route struct {
}
```

Route represents a route record to be used by a Router.

#### type Router

```go
type Router struct {
	NotFoundHandler         http.Handler
	MethodNotAllowedHandler http.Handler
}
```

Router represents a multiplexer that routes HTTP requests.

#### func  NewRouter

```go
func NewRouter() *Router
```
NewRouter constructs and returns a pointer to a new Router.

#### func (*Router) FileHandler

```go
func (r *Router) FileHandler(path string, root http.FileSystem) *Router
```
FileHandler registers a route handler as a file server, serving the directory
qualified as `path`. This method effectively replaces the `Handler()` stage of
the route pipeline. @todo allow usage of custom NotFoundHandler

#### func (*Router) Handler

```go
func (r *Router) Handler(path string, handler http.Handler) *Router
```
Handler adds a path and handler to the current Route record.

#### func (*Router) Register

```go
func (r *Router) Register()
```
Register registers the current Route record. This method must be invoked to
register the Route.

#### func (*Router) ServeHTTP

```go
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
ServeHTTP routes an HTTP request to the appropriate Route record handler.

#### func (*Router) Use

```go
func (r *Router) Use(mws ...middleware) *Router
```
Use adds middlewares to the current Route record.

#### func (*Router) WithMethods

```go
func (r *Router) WithMethods(methods ...string) *Router
```
WithMethods appends user-specified HTTP methods to the current Route record.
