package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var recentVolume = []int64{0, 0, 0, 0, 0}
var base, accent string

func main() {
	flag.StringVar(&base, "base", "#000000", "base color")
	flag.StringVar(&accent, "accent", "#000000", "accent color")
	flag.Parse()

	ver, _ := json.Marshal(version{1})
	fmt.Printf("%s\n[\n[]\n", ver)

	gMuted, _ := json.Marshal(NewGoodPanel("muted", ""))
	gXwayland, _ := json.Marshal(NewGoodPanel("xwayland", ""))
	gVolume, _ := json.Marshal(NewGoodPanel("volume", ""))
	var lock sync.Mutex

	go func() {
		for {
			lMuted, _ := json.Marshal(muted())
			lVolume, _ := json.Marshal(volume())

			lock.Lock()
			gMuted = lMuted
			gVolume = lVolume
			lock.Unlock()

			time.Sleep(1000 * time.Millisecond)
		}
	}()

	go func() {
		for {
			lXwayland, _ := json.Marshal(xwayland())

			lock.Lock()
			gXwayland = lXwayland
			lock.Unlock()

			time.Sleep(1 * time.Minute)
		}
	}()

	var gDate []byte
	for {
		gDate, _ = json.Marshal(date())
		lock.Lock()
		fmt.Printf(",[%s,%s,%s,%s]\n", gXwayland, gMuted, gVolume, gDate)
		lock.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

func date() panel {
	res := time.Now().Format("[Mon] 2006-01-02 15:04:05")

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

	return NewGoodPanel("volume", fmt.Sprintf(" VOL: %d%% ", vol))
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
