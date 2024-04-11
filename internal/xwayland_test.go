package internal

import (
	"testing"
)

func BenchmarkGetXWayland(b *testing.B) {
	var text []byte
	var ok bool
	for i := 0; i < b.N; i++ {
		text, ok = GetXWayland()
		if !ok {
			b.Fatalf("%s, %v\n", text, ok)
		}
	}
}
