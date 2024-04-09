package main

import (
	"log"
	"os"
	"testing"
)

var NullFile *os.File

func init() {
	var err error
	NullFile, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
}

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
