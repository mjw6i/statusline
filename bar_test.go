package main

import (
	"testing"
)

func BenchmarkBarRenderHeader(b *testing.B) {
	bar := NewBar()

	for i := 0; i < b.N; i++ {
		bar.RenderHeader()
	}
}

func BenchmarkBarRenderAll(b *testing.B) {
	bar := NewBar()
	bar.UpdateAll()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.Render()
	}
}
