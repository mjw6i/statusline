package main

import "testing"

func BenchmarkGetSinks(b *testing.B) {
	var vol int64
	var err error
	for i := 0; i < b.N; i++ {
		vol, err = getSinks()
		if err != nil || vol != 40 {
			b.Fatalf("%v, %v\n", vol, err)
		}
	}
}
