package main

import "testing"

func BenchmarkGetIP(b *testing.B) {
	ip := IP{}
	var p panel
	for i := 0; i < b.N; i++ {
		p = ip.GetListeningIP()
		if p.Text != "" {
			b.Fatalf("%v\n", p.Text)
		}
	}
}
