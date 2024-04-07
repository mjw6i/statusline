package main

import (
	"io"
	"slices"
	"testing"
)

func BenchmarkGetSinks(b *testing.B) {
	var vol int
	var err error

	s := Sound{}

	for i := 0; i < b.N; i++ {
		vol, err = s.GetSinks()
		if err != nil || vol != 40 {
			b.Fatalf("%v, %v\n", vol, err)
		}
	}
}

func BenchmarkGetSources(b *testing.B) {
	var p panel

	s := Sound{}

	for i := 0; i < b.N; i++ {
		p = s.GetSources()

		if !slices.Equal(p.Text, []byte(`""`)) {
			b.Fatalf("%+v\n", p)
		}
	}
}

func BenchmarkEventLoop(b *testing.B) {
	data := []byte(`
{"index":39686,"event":"new","on":"client"}
{"index":72,"event":"change","on":"sink"}
{"index":52,"event":"change","on":"card"}
{"index":39686,"event":"remove","on":"client"}
{"index":39687,"event":"new","on":"client"}
{"index":39687,"event":"remove","on":"client"}
{"index":39688,"event":"new","on":"client"}
{"index":72,"event":"change","on":"sink"}
{"index":39688,"event":"remove","on":"client"}
{"index":39689,"event":"new","on":"client"}
{"index":39689,"event":"remove","on":"client"}`)

	updateSources := make(chan struct{})
	updateSinks := make(chan struct{})

	r := newLoopReader(data, b.N)

	eventLoop(r, updateSources, updateSinks)
}

type loopReader struct {
	data    []byte
	times   int
	eof     int
	counter int
}

func newLoopReader(data []byte, times int) io.Reader {
	return &loopReader{data: data, times: times}
}

func (r *loopReader) Read(b []byte) (n int, err error) {
	if r.times == 0 {
		return 0, io.EOF
	}

	n = min(len(r.data)-r.counter, cap(b))

	for i := 0; i < n; i++ {
		b[i] = r.data[r.counter+i]
	}
	r.counter += n
	if r.counter == len(r.data) {
		r.counter = 0
		r.times--
	}

	return n, nil
}

func TestLoopReaderWhole(t *testing.T) {
	r := newLoopReader([]byte("ABC"), 3)
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(b, []byte("ABCABCABC")) {
		t.Fatalf("got: %v\n", b)
	}
}

func TestLoopReaderPartial(t *testing.T) {
	r := newLoopReader([]byte("ABC"), 2)

	var n int
	var err error
	b := make([]byte, 2)

	n, err = r.Read(b)
	assertReader(t, nil, err, 2, n, []byte("AB"), b[:n])
	n, err = r.Read(b)
	assertReader(t, nil, err, 1, n, []byte("C"), b[:n])
	n, err = r.Read(b)
	assertReader(t, nil, err, 2, n, []byte("AB"), b[:n])
	n, err = r.Read(b)
	assertReader(t, nil, err, 1, n, []byte("C"), b[:n])
	n, err = r.Read(b)
	assertReader(t, io.EOF, err, 0, n, []byte(""), b[:n])
}

func assertReader(t *testing.T, expectedErr, actualErr error, expectedCount, actualCount int, expectedBytes, actualBytes []byte) {
	if expectedErr != actualErr {
		t.Fatal(actualErr)
	}
	if expectedCount != actualCount {
		t.Fatalf("got: %v\n", actualCount)
	}
	if !slices.Equal(expectedBytes, actualBytes) {
		t.Fatalf("got: %v\n", actualBytes)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
