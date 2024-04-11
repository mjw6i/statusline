package internal

import (
	"slices"
	"testing"
)

func BenchmarkGetIP(b *testing.B) {
	ip := IP{}
	var p panel
	for i := 0; i < b.N; i++ {
		p = ip.GetListeningIP()
		if !slices.Equal(p.Text, []byte(`""`)) {
			b.Fatalf("%v\n", p.Text)
		}
	}
}
