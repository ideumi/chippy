/*
 *
 * Chippy - internal/optional/regex/cache.go
 *
 */

package regex

import (
	"regexp"
	"sync"
)

/*
 * NOTE: This is intentionally not isolated by actor because there is nothing that
 * can break according to Go docs, and the same cached data shared across
 * multiple actors saves memory and speeds up everything.
 *
 * See: https://pkg.go.dev/regexp#Regexp
 */

const expressionCacheSize = 128

type expressionCache struct {
	mu    sync.Mutex
	items map[string]*regexp.Regexp
	order []string
}

var compiledExpressions = &expressionCache{
	items: make(map[string]*regexp.Regexp, expressionCacheSize),
}

func (cache *expressionCache) get(expression string) (*regexp.Regexp, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	re, hit := cache.items[expression]

	if hit {
		return re, true
	}

	return nil, false
}

func (cache *expressionCache) put(expression string, re *regexp.Regexp) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, hit := cache.items[expression]; hit {
		return
	}

	// FIFO
	if len(cache.items) >= expressionCacheSize {
		oldest := cache.order[0]

		cache.order = cache.order[1:]
		delete(cache.items, oldest)
	}

	cache.items[expression] = re
	cache.order = append(cache.order, expression)
}

func getCompiled(expression string, admitted bool) (*regexp.Regexp, error) {
	if !admitted {
		return regexp.Compile(expression)
	}

	if re, hit := compiledExpressions.get(expression); hit {
		return re, nil
	}

	re, err := regexp.Compile(expression)

	if err != nil {
		return nil, err
	}

	compiledExpressions.put(expression, re)

	return re, nil
}
