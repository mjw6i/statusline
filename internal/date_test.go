package internal

import (
	"testing"
)

func BenchmarkGetDate(b *testing.B) {
	d := NewDate()

	for i := 0; i < b.N; i++ {
		d.getDate()
		if len(d.buf) != 27 {
			b.Fatalf("%s %v\n", d.buf, len(d.buf))
		}
		d.buf = d.buf[:0]
	}
}
