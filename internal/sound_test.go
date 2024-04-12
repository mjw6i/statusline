package internal

import (
	"context"
	"testing"
	"time"
)

func BenchmarkGetSinks(b *testing.B) {
	var vol int
	var ok bool

	s := Sound{}

	for i := 0; i < b.N; i++ {
		vol, ok = s.GetSinks()
		if !ok || vol != 40 {
			b.Fatalf("%v, %v\n", vol, ok)
		}
	}
}

func BenchmarkGetSources(b *testing.B) {
	var text []byte
	var ok bool

	s := Sound{}

	for i := 0; i < b.N; i++ {
		text, ok = s.GetSources()

		if !ok {
			b.Fatalf("%s, %v\n", text, ok)
		}
	}
}

func BenchmarkSubscribe(b *testing.B) {
	updateMic := make(chan struct{})
	updateVolume := make(chan struct{})
	buf := make([]byte, 4096)

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(3 * time.Millisecond)
			cancel()
		}()

		Subscribe(ctx, buf, updateMic, updateVolume)
	}
}

func BenchmarkEventLine(b *testing.B) {
	data := [][]byte{
		[]byte(`{"index":39686,"event":"new","on":"client"}`),
		[]byte(`{"index":72,"event":"change","on":"sink"}`),
		[]byte(`{"index":52,"event":"change","on":"card"}`),
		[]byte(`{"index":39686,"event":"remove","on":"client"}`),
		[]byte(`{"index":39687,"event":"new","on":"client"}`),
		[]byte(`{"index":39687,"event":"remove","on":"client"}`),
		[]byte(`{"index":39688,"event":"new","on":"client"}`),
		[]byte(`{"index":72,"event":"change","on":"sink"}`),
		[]byte(`{"index":39688,"event":"remove","on":"client"}`),
		[]byte(`{"index":39689,"event":"new","on":"client"}`),
		[]byte(`{"index":39689,"event":"remove","on":"client"}`),
	}

	updateSources := make(chan struct{})
	updateSinks := make(chan struct{})

	for i := 0; i < b.N; i++ {
		for _, line := range data {
			eventLine(line, updateSources, updateSinks)
		}
	}
}
