package utilsw

import (
	"github.com/grewwc/go_tools/src/cw"
)

var _cache_map *cw.ConcurrentHashMap[func(...any) any, *cw.LruCache] = cw.NewConcurrentHashMap[func(...any) any, *cw.LruCache](
	nil, nil,
)

func WithCache[T any](maxSize uint, key any, f func(...any) any, args ...any) T {
	cache := _cache_map.GetOrDefault(f, cw.NewLruCache(maxSize))
	_cache_map.PutIfAbsent(f, cache)
	prev := cache.Get(key)
	if prev == nil {
		res := f(args...)
		cache.Put(key, res)
		return res.(T)
	}

	return prev.(T)
}
