package main

import (
	"encoding/json"
	"errors"
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

var recentVolume = []int64{0, 0, 0, 0, 0}

func main() {
	var base, accent string
	flag.StringVar(&base, "base", "#000000", "base color")
	flag.StringVar(&accent, "accent", "#000000", "accent color")
	flag.Parse()

	ver, _ := json.Marshal(version{1})
	fmt.Printf("%s\n[\n[]\n", ver)

	for {
		date, _ := json.Marshal(date())
		muted, _ := json.Marshal(muted())
		xwayland, _ := json.Marshal(xwayland())
		volume, _ := json.Marshal(volume())
		fmt.Printf(",[%s%s%s%s]\n", xwayland, muted, volume, date)
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

func muted() panel {
	out, err := exec.Command("pactl", "get-source-mute", "@DEFAULT_SOURCE@").Output()
	if err != nil {
		return NewBadPanel("muted", "error")
	}

	res := string(out)
	res = strings.TrimSuffix(res, "\n")

	if res == "Mute: yes" {
		return NewGoodPanel("muted", "")
	} else if res == "Mute: no" {
		return NewBadPanel("muted", " not muted ")
	}

	return NewBadPanel("muted", "error")
}

func readVolume() (int64, error) {
	out, err := exec.Command("pactl", "get-sink-volume", "@DEFAULT_SINK@").Output()
	if err != nil {
		return 0, err
	}

	format := "Volume: front-left: %d / %d%% / %f dB, front-right: %d / %d%% / %f dB \n balance %f"

	var fla, flp, fra, frp int64
	var flo, fro, balance float64

	_, err = fmt.Sscanf(string(out), format, &fla, &flp, &flo, &fra, &frp, &fro, &balance)
	if err != nil {
		return 0, err
	}

	return flp, nil
}

func volume() panel {
	vol, err := readVolume()
	if err != nil {
		return NewBadPanel("volume", "error")
	}

	recentVolume = recentVolume[1:]
	recentVolume = append(recentVolume, vol)

	f := recentVolume[0]
	c := false
	for _, v := range recentVolume[1:] {
		if f != v {
			c = true
			break
		}
	}

	if !c {
		return NewGoodPanel("volume", "")
	}

	return NewGoodPanel("volume", fmt.Sprintf("VOL: %d%%", vol))
}

func xwayland() panel {
	err := exec.Command("pidof", "Xwayland").Run()
	if err != nil {
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			ec := eerr.ProcessState.ExitCode()
			if ec == 1 {
				return NewGoodPanel("xwayland", "")
			}
		}
		return NewBadPanel("xwayland", "error")
	}

	return NewBadPanel("xwayland", " xwayland ")
}
