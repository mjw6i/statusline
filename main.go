package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"
)

type version struct {
	Version int `json:"version"`
}

type panel struct {
	Name string `json:"name"`
	Text string `json:"full_text"`
}

func main() {
	var base, accent string
	flag.StringVar(&base, "base", "#000000", "base color")
	flag.StringVar(&accent, "accent", "#000000", "accent color")
	flag.Parse()

	ver, _ := json.Marshal(version{1})
	fmt.Printf("%s\n[\n[]\n", ver)

	for {
		date, _ := json.Marshal(date())
		fmt.Printf(",[%s]\n", date)
		time.Sleep(3 * 1000 * time.Millisecond)
	}
}

func date() panel {
	return panel{"date", "hello"}
}
