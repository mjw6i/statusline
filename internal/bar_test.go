package internal

import (
	"testing"
)

func BenchmarkBarRenderHeader(b *testing.B) {
	bar := NewBar(NullFile)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.RenderHeader()
	}
}

func BenchmarkBarRenderAll(b *testing.B) {
	bar := NewBar(NullFile)
	bar.UpdateAll()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.Render()
	}
}

func BenchmarkUpdateAll(b *testing.B) {
	bar := NewBar(NullFile)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.UpdateAll()
	}
}
