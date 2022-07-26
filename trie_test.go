package turnpike

// @todo refactor: reusability, setup/teardown
import (
	"net/http"
	"reflect"
	"testing"
)

type routeRecord struct {
	path        string
	methods     []string
	handler     http.Handler
	middlewares middlewares
}

func TestNewTrie(t *testing.T) {
	actual := newTrie()
	expected := &trie{
		root: &node{
			children: make(map[string]*node),
			actions:  make(map[string]*action),
		}}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v but got %v\n", actual, expected)
	}
}

func TestInsert(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	records := []routeRecord{
		{
			path:        PathRoot,
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first, second, third},
		},
		{
			path:        PathRoot,
			methods:     []string{http.MethodGet, http.MethodPost},
			handler:     testHandler,
			middlewares: []middleware{first, second, third},
		},
		{
			path:        "/test",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first, second, third},
		},
		{
			path:        "/test/path",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first, second, third},
		},
		{
			path:        "/test/path",
			methods:     []string{http.MethodPost},
			handler:     testHandler,
			middlewares: []middleware{first, second, third},
		},
		{
			path:        "/test/path/paths",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first, second, third},
		},
		{
			path:        "/foo/bar",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first, second, third},
		},
	}

	trie := newTrie()

	for i, record := range records {
		if err := trie.insert(record.methods, record.path, record.handler, record.middlewares); err != nil {
			t.Errorf("error %v inserting test %d\n", err, i)
		}
	}
}

func TestSearchResults(t *testing.T) {
	type searchQuery struct {
		method string
		path   string
	}

	type testCase struct {
		name     string
		search   searchQuery
		expected result
	}

	rootHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	testPathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	testPathPathsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	testPathIdHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	fooHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	barIdHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wildcardHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	insert := []routeRecord{
		{
			path:        PathRoot,
			methods:     []string{http.MethodGet},
			handler:     rootHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path",
			methods:     []string{http.MethodGet},
			handler:     testPathHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path",
			methods:     []string{http.MethodPost},
			handler:     testPathHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path/paths",
			methods:     []string{http.MethodGet},
			handler:     testPathPathsHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path/:id[^\\d+$]",
			methods:     []string{http.MethodGet},
			handler:     testPathIdHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/foo",
			methods:     []string{http.MethodGet},
			handler:     fooHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/bar/:id[^\\d+$]/:user[^\\D+$]",
			methods:     []string{http.MethodPost},
			handler:     barIdHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/:*[(.+)]",
			methods:     []string{http.MethodOptions},
			handler:     wildcardHandler,
			middlewares: []middleware{first},
		},
	}

	tests := []testCase{
		{
			name: "SearchRoot",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/",
			},
			expected: result{
				actions: &action{
					handler:     rootHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{},
			},
		},
		{
			name: "SearchTrailingPath",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/test/",
			},
			expected: result{
				actions: &action{
					handler:     testHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{},
			},
		},
		{
			name: "SearchWithParams",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/test/path/12",
			},
			expected: result{
				actions: &action{
					handler:     testPathIdHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{{
					key:   "id",
					value: "12",
				}},
			},
		},
		{
			name: "SearchNestedPath",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/test/path/paths",
			},
			expected: result{
				actions: &action{
					handler:     testPathPathsHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{},
			},
		},
		{
			name: "SearchPartialPath",
			search: searchQuery{
				method: http.MethodPost,
				path:   "/test/path",
			},
			expected: result{
				actions: &action{
					handler:     testPathHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{},
			},
		},
		{
			name: "SearchPartialPathOtherMethod",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/test/path",
			},
			expected: result{
				actions: &action{
					handler:     testPathHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{},
			},
		},
		{
			name: "SearchAdditionalBasePath",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/foo",
			},
			expected: result{
				actions: &action{
					handler:     fooHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{},
			},
		},
		{
			name: "SearchAdditionalBasePathTrailingSlash",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/foo/",
			},
			expected: result{
				actions: &action{
					handler:     fooHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{},
			},
		},
		{
			name: "SearchComplexRegex",
			search: searchQuery{
				method: http.MethodPost,
				path:   "/bar/123/alice",
			},
			expected: result{
				actions: &action{
					handler:     barIdHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{
					{
						key:   "id",
						value: "123",
					},
					{
						key:   "user",
						value: "alice",
					},
				},
			},
		},
		{
			name: "SearchWildcardRegex",
			search: searchQuery{
				method: http.MethodOptions,
				path:   "/wildcard",
			},
			expected: result{
				actions: &action{
					handler:     wildcardHandler,
					middlewares: []middleware{first},
				},
				parameters: []*parameter{
					{
						key:   "*",
						value: "wildcard",
					},
				},
			},
		},
	}

	trie := newTrie()

	for _, record := range insert {
		trie.insert(record.methods, record.path, record.handler, record.middlewares)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := trie.search(test.search.method, test.search.path)

			if err != nil {
				t.Errorf("expected a result but got error %v", err)
			}

			if reflect.ValueOf(actual.actions.handler) != reflect.ValueOf(test.expected.actions.handler) {
				t.Errorf("expected %v but got %v", test.expected.actions.handler, actual.actions.handler)
			}

			if len(actual.parameters) != len(test.expected.parameters) {
				t.Errorf("expected %v but got %v", len(test.expected.parameters), len(actual.parameters))
			}

			for i, param := range actual.parameters {
				if !reflect.DeepEqual(param, test.expected.parameters[i]) {
					t.Errorf("expected %v but got %v", test.expected.parameters[i], param)
				}
			}
		})
	}
}

func TestSearchError(t *testing.T) {
	type searchQuery struct {
		method string
		path   string
	}

	type testCase struct {
		name   string
		search searchQuery
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	insert := []routeRecord{
		{
			path:        PathRoot,
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first},
		},
		{
			path:        PathRoot,
			methods:     []string{http.MethodGet, http.MethodPost},
			handler:     testHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path",
			methods:     []string{http.MethodPost},
			handler:     testHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path/paths",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first},
		},
		{
			path:        "/test/path/:id[^\\d+$]",
			methods:     []string{http.MethodGet},
			handler:     testHandler,
			middlewares: []middleware{first},
		}}

	tests := []testCase{
		{
			name: "SearchComplexRegex",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/test/path/12/31",
			},
		},
		{
			name: "SearchNestedPath",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/test/path/path",
			},
		},
		{
			name: "SearchSpaceInPath",
			search: searchQuery{
				method: http.MethodPost,
				path:   "/test/pat h",
			},
		},
		{
			name: "SearchNestedPathAlt",
			search: searchQuery{
				method: http.MethodGet,
				path:   "/test/path/world",
			},
		},
	}

	trie := newTrie()

	for _, record := range insert {
		trie.insert(record.methods, record.path, record.handler, record.middlewares)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := trie.search(test.search.method, test.search.path)

			if err == nil {
				t.Errorf("expected an error but got result %v", result)
			}
		})
	}
}
