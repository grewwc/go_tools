package utilsw

import (
	"github.com/grewwc/go_tools/src/cw"
)

var _cache_map *cw.ConcurrentHashMap[func(...any) any, *cw.LruCache] = cw.NewConcurrentHashMap[func(...any) any, *cw.LruCache](
	nil, nil,
)

func WithCache[returnType any](cacheSize uint, cacheKey any, f func(...any) any, args ...any) returnType {
	cache := _cache_map.GetOrDefault(f, cw.NewLruCache(cacheSize))
	_cache_map.PutIfAbsent(f, cache)
	prev := cache.Get(cacheKey)
	if prev == nil {
		res := f(args...)
		cache.Put(cacheKey, res)
		return res.(returnType)
	}

	return prev.(returnType)
}
