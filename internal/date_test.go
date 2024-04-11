package internal

import (
	"slices"
	"testing"
)

func BenchmarkGetDate(b *testing.B) {
	d := NewDate()
	var p panel

	for i := 0; i < b.N; i++ {
		p = d.GetDate()
		if slices.Equal(p.Text, []byte(`""`)) {
			b.Fatalf("%+v\n", p)
		}
	}
}
