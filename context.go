package turnpike

import "context"

type key int

const (
	// parameterKey is a request context key.
	parameterKey key = iota
)

// GetParam retrieves from context a value corresponding to a given key.
func GetParam(ctx context.Context, key string) string {
	params, _ := ctx.Value(parameterKey).([]*parameter)

	for i := range params {
		if params[i].key == key {
			return params[i].value
		}
	}

	return ""
}
