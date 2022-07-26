package turnpike

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestNewMiddleware(t *testing.T) {
	actual := newMiddlewares(nil)
	var expected middlewares

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v but got %v\n", actual, expected)
	}
}

func TestMiddlewareInvocation(t *testing.T) {
	r := NewRouter()

	r.WithMethods(http.MethodGet).Handler("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "/")
	})).Use(first, second, third).Register()

	r.WithMethods(http.MethodPost).Handler("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "foo")
	})).Use(first, second).Register()

	r.WithMethods(http.MethodPost).Handler("/foo/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "foo")
	})).Use(func(next http.Handler) http.Handler {
		// Test access to context params
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "first: before\n")

			id := GetParam(r.Context(), "id")
			fmt.Fprintf(w, "%v\n", id)

			url, _ := url.Parse(r.URL.RawQuery)
			values := url.Query()

			values.Add("test", "testval")
			url.RawQuery = values.Encode()

			next.ServeHTTP(w, r)
			fmt.Fprintf(w, "first: after\n")
		})
	}, second).Register()

	tests := []testCase{
		{
			name:   "Middleware",
			path:   "/",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "first: before\nsecond: before\nthird: before\n/third: after\nsecond: after\nfirst: after\n",
		},
		{
			name:   "MiddlewareAltMethod",
			path:   "/foo",
			method: http.MethodPost,
			code:   http.StatusOK,
			body:   "first: before\nsecond: before\nfoosecond: after\nfirst: after\n",
		},
		{
			name:   "MiddlewareWithContext",
			path:   "/foo/20",
			method: http.MethodPost,
			code:   http.StatusOK,
			body:   "first: before\n20\nsecond: before\nfoosecond: after\nfirst: after\n",
		},
	}

	runHTTPTests(t, r, tests)
}

func first(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "first: before\n")
		next.ServeHTTP(w, r)
		fmt.Fprintf(w, "first: after\n")
	})
}

func second(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "second: before\n")
		next.ServeHTTP(w, r)
		fmt.Fprintf(w, "second: after\n")
	})
}

func third(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "third: before\n")
		next.ServeHTTP(w, r)
		fmt.Fprintf(w, "third: after\n")
	})
}
