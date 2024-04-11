package internal

import (
	"testing"
)

func BenchmarkGetIP(b *testing.B) {
	ip := IP{}
	var text []byte
	var ok bool
	for i := 0; i < b.N; i++ {
		text, ok = ip.GetListeningIP()
		if !ok {
			b.Fatalf("%s, %v\n", text, ok)
		}
	}
}
