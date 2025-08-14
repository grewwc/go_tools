package cw

type node struct {
	key   interface{}
	value interface{}
	next  *node
	prev  *node
}

type LruCache struct {
	m     map[interface{}]*node
	head  *node
	tail  *node
	cap   uint
	count uint
}

func NewLruCache(cap uint) *LruCache {
	ret := LruCache{cap: cap, m: make(map[interface{}]*node)}
	ret.head = &node{}
	ret.tail = &node{}
	ret.head.next = ret.tail
	ret.tail.prev = ret.head
	return &ret
}

func (cache *LruCache) Get(key interface{}) interface{} {
	n, ok := cache.m[key]
	if !ok {
		return nil
	}
	cache.moveToFront(n)
	return n.value
}

func (cache *LruCache) Put(key interface{}, value interface{}) {
	n, ok := cache.m[key]
	if !ok {
		newNode := &node{key: key, value: value}
		cache.m[key] = newNode
		cache.insertToFront(newNode)
		cache.count++
		if cache.count > cache.cap {
			cache.removeTail()
		}

	} else {
		n.value = value
		cache.moveToFront(n)
	}
}

func (cache *LruCache) insertToFront(n *node) {
	front := cache.head.next

	cache.head.next = n
	n.prev = cache.head

	n.next = front
	front.prev = n
}

func (cache *LruCache) moveToFront(n *node) {
	next := n.next
	prev := n.prev
	next.prev = prev
	prev.next = next

	front := cache.head.next
	cache.head.next = n
	n.prev = cache.head

	n.next = front
	front.prev = n
}

func (cache *LruCache) removeTail() {
	tail := cache.tail.prev
	tail.prev.next = tail
	cache.tail.prev = tail.prev
	delete(cache.m, tail.key)
	tail = nil
}
