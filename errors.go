package turnpike

import "errors"

var (
	ErrNotFound         = errors.New("no matching route record found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)
