package cache

import (
	"sync"
)

// entry represents a cache item in the doubly-linked list.
type entry[K comparable, V any] struct {
	key  K
	val  V
	prev *entry[K, V]
	next *entry[K, V]
}

// LRU is a generic, thread-safe least-recently-used cache.
type LRU[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	items    map[K]*entry[K, V]
	head     *entry[K, V] // most recent
	tail     *entry[K, V] // least recent
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity < 1 {
		capacity = 128
	}
	return &LRU[K, V]{
		capacity: capacity,
		items:    make(map[K]*entry[K, V], capacity),
	}
}

// Get retrieves a value and moves it to the front (most recent).
func (c *LRU[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.moveToFront(e)
	return e.val, true
}

// Put inserts or updates a key-value pair. Evicts LRU if at capacity.
func (c *LRU[K, V]) Put(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.items[key]; ok {
		e.val = val
		c.moveToFront(e)
		return
	}

	e := &entry[K, V]{key: key, val: val}
	c.items[key] = e
	c.pushFront(e)

	if len(c.items) > c.capacity {
		c.evict()
	}
}

// Delete removes a key from the cache.
func (c *LRU[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		return
	}
	c.remove(e)
	delete(c.items, key)
}

// Len returns the current number of items.
func (c *LRU[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// Clear removes all items.
func (c *LRU[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*entry[K, V], c.capacity)
	c.head = nil
	c.tail = nil
}

func (c *LRU[K, V]) pushFront(e *entry[K, V]) {
	e.prev = nil
	e.next = c.head
	if c.head != nil {
		c.head.prev = e
	}
	c.head = e
	if c.tail == nil {
		c.tail = e
	}
}

func (c *LRU[K, V]) moveToFront(e *entry[K, V]) {
	if c.head == e {
		return
	}
	c.remove(e)
	c.pushFront(e)
}

func (c *LRU[K, V]) remove(e *entry[K, V]) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.tail = e.prev
	}
}

func (c *LRU[K, V]) evict() {
	if c.tail == nil {
		return
	}
	delete(c.items, c.tail.key)
	c.remove(c.tail)
}
