package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/cw"
)

func TestLruCache(t *testing.T) {
	c := cw.NewLruCache(2)
	c.Put("k1", 1)
	c.Put("k2", 2)
	if c.Get("k1").(int) != 1 {
		t.Fatal("error")
	}
	c.Put("k3", 3)
	if c.Get("k2") != nil {
		t.Fatal("k2 should be nil")
	}
	c.Get("k2")
	c.Get("k3")
	if c.Get("k1").(int) != 1 {
		t.Fatal("k1 should be 1")
	}
	c.Get("k4")
	c.Get("k3")
	c.Put("k4", 4)
	if c.Get("k4") != 4 {
		t.Fatal("k4 should be 4")
	}
	if c.Get("k1") != nil {
		t.Fatal("k1 should be nil")
	}
}
