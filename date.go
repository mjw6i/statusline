package main

import "time"

func date() panel {
	res := time.Now().Format("[Mon] 2006-01-02 15:04:05")

	return NewGoodPanel("date", res)
}
