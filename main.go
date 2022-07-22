package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type version struct {
	Version int `json:"version"`
}

type panel struct {
	Name string `json:"name"`
	Text string `json:"full_text"`
}

func NewGoodPanel(name string, text string) panel {
	return panel{name, text}
}

func NewBadPanel(name string, text string) panel {
	return panel{name, text}
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
	out, err := exec.Command("date", "+[%a] %Y-%m-%d %H:%M:%S").Output()
	if err != nil {
		return NewBadPanel("date", "error")
	}

	res := string(out)
	res = strings.TrimSuffix(res, "\n")

	return NewGoodPanel("date", res)
}
