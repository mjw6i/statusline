package main

const keep = 5

type vol struct {
	val   int64
	times int
}

func (v *vol) Push(val int64) {
	if v.val != val {
		v.val = val
		v.times = 0
		return
	}

	if v.times < keep {
		v.times++
	}
}

func (v *vol) Same() bool {
	return v.times == keep
}
