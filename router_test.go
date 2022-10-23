package turnpike

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type testCase struct {
	name   string
	path   string
	method string
	code   int
	body   string
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("")
}

func TestNewRouter(t *testing.T) {
	actual := NewRouter()
	expected := &Router{
		trie: newTrie(),
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v but got %v\n", actual, expected)
	}
}

func TestRouteHandler(t *testing.T) {
	r := NewRouter()

	r.WithMethods(http.MethodGet).Handler("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "/")
	})).Register()

	r.WithMethods(http.MethodGet).Handler("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "foo")
	})).Register()

	r.WithMethods(http.MethodGet).Handler("/foo/bar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "foobar")
	})).Register()

	r.WithMethods(http.MethodGet).Handler("/foo/bar/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r.Context(), "id")
		fmt.Fprintf(w, "/foo/bar/%v", id)
	})).Register()

	r.Handler("/baz/:id/:user", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r.Context(), "id")
		user := GetParam(r.Context(), "user")
		fmt.Fprintf(w, "/baz/%v/%v", id, user)
	})).WithMethods(http.MethodGet).Register()

	r.Handler("/foo/:id[^\\d+$]", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r.Context(), "id")
		fmt.Fprintf(w, "/foo/%v", id)
	})).WithMethods(http.MethodDelete).Register()

	r.WithMethods(http.MethodOptions).Handler("/options", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "/options")
	})).Register()

	tests := []testCase{
		{
			name:   "RootPath",
			path:   "/",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "/",
		},
		{
			name:   "BasicPath",
			path:   "/foo",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "foo",
		},
		{
			name:   "NestedPath",
			path:   "/foo/bar",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "foobar",
		},
		{
			name:   "PathWithParams",
			path:   "/foo/bar/123",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "/foo/bar/123",
		},
		{
			name:   "PathWithComplexParams",
			path:   "/baz/123/bob",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "/baz/123/bob",
		},
		{
			name:   "PartialPathWithParams",
			path:   "/foo/21",
			method: http.MethodDelete,
			code:   http.StatusOK,
			body:   "/foo/21",
		},
		{
			name:   "OptionsPath",
			path:   "/options",
			method: http.MethodOptions,
			code:   http.StatusOK,
			body:   "/options",
		},
	}

	runHTTPTests(t, r, tests)
}

func TestMultiMethodRouteHandler(t *testing.T) {
	r := NewRouter()

	r.WithMethods(http.MethodGet, http.MethodPost).Handler("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "/")
	})).Register()

	// Test registration of methods via multiple WithMethods invocations
	r.WithMethods(http.MethodGet).Handler("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "foo")
	})).WithMethods(http.MethodPost).Register()

	r.WithMethods(http.MethodGet, http.MethodPost).Handler("/foo/bar/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r.Context(), "id")
		fmt.Fprintf(w, "/foo/bar/%v", id)
	})).Register()

	r.Handler("/baz/:id/:user", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r.Context(), "id")
		user := GetParam(r.Context(), "user")
		fmt.Fprintf(w, "/baz/%v/%v", id, user)
	})).WithMethods(http.MethodGet, http.MethodPost).Register()

	r.Handler("/foo/:id[^\\d+$]", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r.Context(), "id")
		fmt.Fprintf(w, "/foo/%v", id)
	})).WithMethods(http.MethodPost, http.MethodDelete).Register()

	tests := []testCase{
		{
			name:   "RootPathGET",
			path:   "/",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "/",
		},
		{
			name:   "RootPathPOST",
			path:   "/",
			method: http.MethodPost,
			code:   http.StatusOK,
			body:   "/",
		},
		{
			name:   "BasicPathGET",
			path:   "/foo",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "foo",
		},
		{
			name:   "BasicPathPOST",
			path:   "/foo",
			method: http.MethodPost,
			code:   http.StatusOK,
			body:   "foo",
		},
		{
			name:   "ParamsPathGET",
			path:   "/foo/bar/123",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "/foo/bar/123",
		},
		{
			name:   "ParamsPathPOST",
			path:   "/foo/bar/123",
			method: http.MethodPost,
			code:   http.StatusOK,
			body:   "/foo/bar/123",
		},
		{
			name:   "ComplexParamsPathGET",
			path:   "/baz/123/bob",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "/baz/123/bob",
		},
		{
			name:   "ComplexParamsPathPOST",
			path:   "/baz/123/bob",
			method: http.MethodPost,
			code:   http.StatusOK,
			body:   "/baz/123/bob",
		},
		{
			name:   "AltPathPOST",
			path:   "/foo/21",
			method: http.MethodPost,
			code:   http.StatusOK,
			body:   "/foo/21",
		},
		{
			name:   "AltPathDELETE",
			path:   "/foo/21",
			method: http.MethodDelete,
			code:   http.StatusOK,
			body:   "/foo/21",
		},
	}

	runHTTPTests(t, r, tests)
}

func TestDefaultErrorHandlers(t *testing.T) {
	r := NewRouter()

	r.WithMethods(http.MethodGet).Handler(`/notfound`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).Register()
	r.WithMethods(http.MethodGet).Handler(`/notallowed`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).Register()

	tests := []testCase{
		{
			name:   "DefaultNotFoundHandler",
			path:   "/",
			method: http.MethodGet,
			code:   http.StatusNotFound,
			body:   "",
		},
		{
			name:   "DefaultMethodNotAllowedHandler",
			path:   "/notallowed",
			method: http.MethodPost,
			code:   http.StatusMethodNotAllowed,
			body:   "",
		},
	}

	runHTTPTests(t, r, tests)
}

func TestCustomNotFoundHandler(t *testing.T) {
	r := NewRouter()
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "NotFound")
	})

	tests := []testCase{
		{
			name:   "CustomNotFoundHandler",
			path:   "/",
			method: http.MethodGet,
			code:   http.StatusNotFound,
			body:   "NotFound",
		},
		{
			name:   "CustomNotFoundHandlerAltMethod",
			path:   "/notfound",
			method: http.MethodPost,
			code:   http.StatusNotFound,
			body:   "NotFound",
		},
	}

	runHTTPTests(t, r, tests)
}

func TestCustomMethodNotAllowedHandler(t *testing.T) {
	r := NewRouter()
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "MethodNotAllowed")
	})

	r.WithMethods(http.MethodGet, http.MethodDelete).Handler("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})).Register()

	tests := []testCase{
		{
			name:   "MethodAllowed",
			path:   "/",
			method: http.MethodGet,
			code:   http.StatusOK,
			body:   "OK",
		},
		{
			name:   "MethodAllowedAltMethod",
			path:   "/",
			method: http.MethodDelete,
			code:   http.StatusOK,
			body:   "OK",
		},
		{
			name:   "MethodNotAllowed1",
			path:   "/",
			method: http.MethodPost,
			code:   http.StatusMethodNotAllowed,
			body:   "MethodNotAllowed",
		},
		{
			name:   "MethodNotAllowed2",
			path:   "/",
			method: http.MethodPatch,
			code:   http.StatusMethodNotAllowed,
			body:   "MethodNotAllowed",
		},
		{
			name:   "MethodNotAllowed3",
			path:   "/",
			method: http.MethodPut,
			code:   http.StatusMethodNotAllowed,
			body:   "MethodNotAllowed",
		},
		{
			name:   "MethodNotAllowed4",
			path:   "/",
			method: http.MethodOptions,
			code:   http.StatusMethodNotAllowed,
			body:   "MethodNotAllowed",
		},
	}

	runHTTPTests(t, r, tests)
}

func TestFileHandler(t *testing.T) {
	r := NewRouter()
	mfs := &mockFileSystem{}
	r.FileHandler("/", mfs).Register()
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	if mfs.opened {
		t.Error("serving file failed")
	}

	r.ServeHTTP(rec, req)

	if !mfs.opened {
		t.Error("serving file failed")
	}
}

func TestFileHandlerInvariantViolation(t *testing.T) {
	r := NewRouter()
	mfs := &mockFileSystem{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected an invariant violation panic")
		}
	}()

	r.FileHandler("/", mfs).WithMethods(http.MethodPost).Register()
}

func runHTTPTests(t *testing.T, r *Router, tests []testCase) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != test.code {
				t.Errorf("expected code %d but got %d\n", test.code, rec.Code)
			}

			if test.body != "" {
				bodyBytes, _ := ioutil.ReadAll(rec.Body)
				body := string(bodyBytes)
				if body != test.body {
					t.Errorf("expected body %s but got %s\n", test.body, body)
				}
			}
		})
	}
}
