package cache

import (
	"testing"
)

func TestLRU_PutGet(t *testing.T) {
	c := NewLRU[string, int](3)

	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)

	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf("Get(a) = %d, %v; want 1, true", v, ok)
	}
	if v, ok := c.Get("b"); !ok || v != 2 {
		t.Errorf("Get(b) = %d, %v; want 2, true", v, ok)
	}
}

func TestLRU_Eviction(t *testing.T) {
	c := NewLRU[string, int](2)

	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3) // evicts "a"

	if _, ok := c.Get("a"); ok {
		t.Error("expected a to be evicted")
	}
	if v, ok := c.Get("b"); !ok || v != 2 {
		t.Error("b should still be present")
	}
	if v, ok := c.Get("c"); !ok || v != 3 {
		t.Error("c should be present")
	}
}

func TestLRU_LRUOrder(t *testing.T) {
	c := NewLRU[string, int](2)

	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a") // access a → a becomes most recent, b becomes LRU
	c.Put("c", 3) // evicts b (not a)

	if _, ok := c.Get("b"); ok {
		t.Error("expected b to be evicted (LRU)")
	}
	if _, ok := c.Get("a"); !ok {
		t.Error("a should still be present (recently accessed)")
	}
}

func TestLRU_Update(t *testing.T) {
	c := NewLRU[string, int](2)

	c.Put("a", 1)
	c.Put("a", 10) // update

	if v, ok := c.Get("a"); !ok || v != 10 {
		t.Errorf("expected updated value 10, got %d", v)
	}
	if c.Len() != 1 {
		t.Errorf("expected len 1 after update, got %d", c.Len())
	}
}

func TestLRU_Delete(t *testing.T) {
	c := NewLRU[string, int](3)

	c.Put("a", 1)
	c.Put("b", 2)
	c.Delete("a")

	if _, ok := c.Get("a"); ok {
		t.Error("a should be deleted")
	}
	if c.Len() != 1 {
		t.Errorf("expected len 1, got %d", c.Len())
	}
}

func TestLRU_Clear(t *testing.T) {
	c := NewLRU[string, int](3)

	c.Put("a", 1)
	c.Put("b", 2)
	c.Clear()

	if c.Len() != 0 {
		t.Errorf("expected len 0 after clear, got %d", c.Len())
	}
}

func TestLRU_GetMiss(t *testing.T) {
	c := NewLRU[string, int](3)
	if _, ok := c.Get("nonexistent"); ok {
		t.Error("expected miss for nonexistent key")
	}
}
