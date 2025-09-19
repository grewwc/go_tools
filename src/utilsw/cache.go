package utilsw

import (
	"github.com/grewwc/go_tools/src/cw"
)

var _cache_map *cw.ConcurrentHashMap[func(...any) any, *cw.LruCache] = cw.NewConcurrentHashMap[func(...any) any, *cw.LruCache](
	nil, nil,
)

// Call caches the result of function f with specified cache size and key.
// It uses an LRU cache to store the results, and returns the cached result if available,
// otherwise it executes the function and caches the result.
//
// Parameters:
//   - cacheSize: the maximum number of entries that can be stored in the cache
//   - cacheKey: the key used to identify the cached result
//   - f: the function to be executed and cached
//   - args: variadic arguments to be passed to function f
//
// Returns:
//   - returnType: the result of function f, either from cache or fresh execution
func Call[returnType any](cacheSize uint, cacheKey any, f func(...any) any, args ...any) returnType {
	// Get or create cache for function f
	cache := _cache_map.GetOrDefault(f, cw.NewLruCache(cacheSize))
	_cache_map.PutIfAbsent(f, cache)

	// Try to get cached result
	prev := cache.Get(cacheKey)
	if prev == nil {
		// Execute function and cache result if not found in cache
		res := f(args...)
		cache.Put(cacheKey, res)
		return res.(returnType)
	}

	// Return cached result
	return prev.(returnType)
}
