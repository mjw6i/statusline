package main

import (
	"testing"
)

func TestOverwrite(t *testing.T) {
	r := NewRing()
	r.Push(1)
	r.Push(2)
	r.Push(3)
	r.Push(4)
	r.Push(5)
	r.Push(6)

	expected := [5]int64{6, 2, 3, 4, 5}

	if r.data != expected {
		t.Errorf("Expected: %v, got %v", expected, r.data)
	}
}
