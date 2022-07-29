package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var base, accent string

func init() {
	debug.SetGCPercent(20)
	runtime.GOMAXPROCS(1)
}

func main() {
	flag.StringVar(&base, "base", "#000000", "base color")
	flag.StringVar(&accent, "accent", "#000000", "accent color")
	flag.Parse()

	ver, _ := json.Marshal(version{1})
	fmt.Printf("%s\n[\n[]\n", ver)

	gMuted, _ := json.Marshal(NewGoodPanel("muted", ""))
	gXwayland, _ := json.Marshal(NewGoodPanel("xwayland", ""))
	gVolume, _ := json.Marshal(NewGoodPanel("volume", ""))
	var gDate []byte
	var recentVolume = []int64{0, 0, 0, 0, 0}

	tPulse := time.Now()
	tXwayland := time.Now()

	var t time.Time

	for {
		t = time.Now()

		switch {
		case t.After(tPulse):
			tPulse = t.Add(1 * time.Second)
			gMuted, _ = json.Marshal(muted())
			gVolume, _ = json.Marshal(volume(&recentVolume))
		case t.After(tXwayland):
			tXwayland = t.Add(1 * time.Minute)
			gXwayland, _ = json.Marshal(xwayland())
		}

		gDate, _ = json.Marshal(date())
		fmt.Printf(",[%s,%s,%s,%s]\n", gXwayland, gMuted, gVolume, gDate)

		time.Sleep(time.Until(t.Add(100 * time.Millisecond)))
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

func volume(recentVolume *[]int64) panel {
	vol, err := readVolume()
	if err != nil {
		return NewBadPanel("volume", "error")
	}

	*recentVolume = (*recentVolume)[1:]
	*recentVolume = append(*recentVolume, vol)

	f := (*recentVolume)[0]
	c := false
	for _, v := range (*recentVolume)[1:] {
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
