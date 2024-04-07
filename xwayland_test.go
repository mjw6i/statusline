package main

import "testing"

func BenchmarkGetXWayland(b *testing.B) {
	var p panel
	for i := 0; i < b.N; i++ {
		p = GetXWayland()
		if p.Text != "" {
			b.Fatalf("%v\n", p.Text)
		}
	}
}
