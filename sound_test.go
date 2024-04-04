package main

import "testing"

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
