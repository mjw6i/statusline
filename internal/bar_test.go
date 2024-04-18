package internal

import (
	"encoding/json"
	"fmt"
	"testing"
)

func BenchmarkBarRenderHeader(b *testing.B) {
	bar := NewBar(NullFile, "#000000", "#000000")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.RenderHeader()
	}
}

func BenchmarkBarRenderAll(b *testing.B) {
	bar := NewBar(NullFile, "#000000", "#000000")
	bar.UpdateAll()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.Render()
	}
}

func BenchmarkUpdateAll(b *testing.B) {
	bar := NewBar(NullFile, "#000000", "#000000")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.UpdateAll()
	}
}

func BenchmarkNewUpdateAndRender(b *testing.B) {
	bar := NewBar(NullFile, "#000000", "#000000")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bar.UpdateAll()
		for range 5 {
			bar.Render()
		}
	}
}

func BenchmarkOldUpdateAndRender(b *testing.B) {
	var gMuted []byte
	var err error

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gMuted, err = json.Marshal(GetSources())
		if err != nil {
			panic(err)
		}
		for range 5 {
			fmt.Fprintf(NullFile, ",[%s]\n", gMuted)
		}
	}
}
