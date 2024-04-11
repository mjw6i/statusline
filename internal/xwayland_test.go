package internal

import (
	"slices"
	"testing"
)

func BenchmarkGetXWayland(b *testing.B) {
	var p panel
	for i := 0; i < b.N; i++ {
		p = GetXWayland()
		if !slices.Equal(p.Text, []byte(`""`)) {
			b.Fatalf("%v\n", p.Text)
		}
	}
}
