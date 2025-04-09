package conw

type node struct {
	key   interface{}
	value interface{}
	next  *node
	prev  *node
}

type LruCache struct {
	_map  map[interface{}]*node
	head  *node
	tail  *node
	cap   uint
	count uint
}

func NewLruCache(cap uint) *LruCache {
	ret := LruCache{cap: cap, _map: make(map[interface{}]*node)}
	ret.head = &node{}
	return &ret
}

func (cache *LruCache) Get(key interface{}) interface{} {
	n, ok := cache._map[key]
	if !ok {
		return nil
	}
	cache.moveToHead(n)
	return n.value
}

func (cache *LruCache) Put(key interface{}, value interface{}) {
	n, ok := cache._map[key]
	if !ok {
		newNode := &node{key: key, value: value}
		cache._map[key] = newNode
		cache.moveToHead(newNode)
		cache.count++
		if cache.tail == nil {
			cache.tail = newNode
		}
		if cache.count > cache.cap {
			cache.removeTail()
		}

	} else {
		n.value = value
		cache.moveToHead(n)
	}
}

func (cache *LruCache) moveToHead(n *node) {
	currHead := cache.head.next
	if n == currHead {
		return
	}
	next := n.next
	prev := n.prev
	if next != nil {
		next.prev = prev
	}
	if prev != nil {
		prev.next = next
	}
	cache.head.next = n
	n.prev = cache.head
	n.next = currHead
	if currHead != nil {
		currHead.prev = n
	}
	if n == cache.tail {
		cache.tail = prev
	}
}

func (cache *LruCache) removeTail() {
	if cache.tail == nil {
		return
	}
	prev := cache.tail.prev

	if _, ok := cache._map[cache.tail.key]; ok {
		delete(cache._map, cache.tail.key)
		cache.count--
	}
	if prev != nil {
		prev.next = nil
	}
	cache.tail = prev
}
