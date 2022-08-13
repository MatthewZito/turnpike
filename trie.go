package turnpike

import (
	"net/http"
)

// action represents an HTTP handler action.
type action struct {
	handler     http.Handler
	middlewares middlewares
}

// parameter represents a path parameter.
type parameter struct {
	key   string
	value string
}

// result represents a trie search result.
type result struct {
	actions    *action
	parameters []*parameter
}

// trie is a trie data structure used to manage multiplexing paths.
type trie struct {
	root *node
}

// node is a trie node.
type node struct {
	label    string
	children map[string]*node
	actions  map[string]*action
}

var rc = newCache()

// newResult constructs and returns a pointer to a new result.
func newResult() *result {
	return &result{}
}

// newTrie constructs and returns a pointer to a new trie.
func newTrie() *trie {
	return &trie{
		root: &node{
			children: make(map[string]*node),
			actions:  make(map[string]*action),
		},
	}
}

// insert inserts a new routing result into the trie.
func (t *trie) insert(methods []string, path string, handler http.Handler, mws middlewares) error {
	curr := t.root

	// Handle root path
	if path == PathRoot {
		curr.label = path
		for _, method := range methods {
			curr.actions[method] = &action{
				handler:     handler,
				middlewares: mws,
			}
		}

		return nil
	}

	paths := expandPath(path)
	for i, splitPath := range paths {
		next, ok := curr.children[splitPath]

		if ok {
			curr = next
		} else {
			curr.children[splitPath] = &node{
				label:    splitPath,
				actions:  make(map[string]*action),
				children: make(map[string]*node),
			}
			curr = curr.children[splitPath]
		}

		// Overwrite existing data on last path
		if i == len(paths)-1 {
			curr.label = splitPath
			for _, method := range methods {
				curr.actions[method] = &action{
					handler:     handler,
					middlewares: mws,
				}
			}

			break
		}

	}

	return nil
}

// search searches a given path and method in the trie's routing results.
func (t *trie) search(method string, searchPath string) (*result, error) {
	var params []*parameter
	result := newResult()
	curr := t.root

	for _, path := range expandPath(searchPath) {
		next, ok := curr.children[path]

		if ok {
			curr = next
			continue
		}

		if len(curr.children) == 0 {
			if curr.label != path {
				// No matching route result found.
				return nil, ErrNotFound
			}
			break
		}

		isParamMatch := false
		for child := range curr.children {
			if string([]rune(child)[0]) == ParameterDelimiter {
				pattern := deriveLabelPattern(child)
				regex, err := rc.get(pattern)

				if err != nil {
					return nil, ErrNotFound
				}

				if regex.Match([]byte(path)) {
					param := deriveParameterKey(child)

					params = append(params, &parameter{
						key:   param,
						value: path,
					})

					curr = curr.children[child]

					isParamMatch = true
					break
				}

				// No parameter match.
				return nil, ErrNotFound
			}
		}

		// No parameter match.
		if !isParamMatch {
			return nil, ErrNotFound

		}
	}

	if searchPath == PathRoot {
		// No matching handler.
		if len(curr.actions) == 0 {
			return nil, ErrNotFound
		}
	}

	result.actions = curr.actions[method]

	// No matching handler.
	if result.actions == nil {
		return nil, ErrMethodNotAllowed
	}

	result.parameters = params

	return result, nil
}
