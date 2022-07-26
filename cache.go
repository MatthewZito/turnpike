package turnpike

import (
	"fmt"
	"regexp"
	"sync"
)

// @todo implement eviction strategy e.g. LRU
// regexCache maintains a thread-safe cache for compiled regular expressions.
type regexCache struct {
	state sync.Map
}

// newCache constructs and returns a pointer to a new regexCache.
func newCache() *regexCache {
	cache := &regexCache{
		state: sync.Map{},
	}

	return cache
}

// get retrieves a compiled regex from the regexCache, or creates one and caches it if not extant.
func (rc *regexCache) get(pattern string) (*regexp.Regexp, error) {
	v, ok := rc.state.Load(pattern)
	if ok {
		// Verify the validity of the cached regex.
		regex, ok := v.(*regexp.Regexp)
		if !ok {
			// @todo delete entry?
			return nil, fmt.Errorf("the given pattern %s is not a valid regular expression", pattern)
		}
		// Return the cached regex.
		return regex, nil
	}

	// Compile the regex and add to cache if valid.
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	rc.state.Store(pattern, regex)
	return regex, nil
}
