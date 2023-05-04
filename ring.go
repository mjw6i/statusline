package main

const RingSize = 5

type ring struct {
	head int
	data [RingSize]int64
}

func NewRing() *ring {
	return &ring{}
}

func (r *ring) Push(val int64) {
	r.data[r.head] = val
	r.head = (r.head + 1) % RingSize
}

func (r *ring) Same() bool {
	for i := 1; i < RingSize; i++ {
		if r.data[0] != r.data[i] {
			return false
		}
	}

	return true
}
