package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"
)

var base, accent string

func init() {
	// debug.SetGCPercent(20)
	runtime.GOMAXPROCS(1)
}

func main() {
	flag.StringVar(&base, "base", "#000000", "base color")
	flag.StringVar(&accent, "accent", "#000000", "accent color")
	flag.Parse()

	updateMic := make(chan struct{}, 1)
	updateVolume := make(chan struct{}, 1)
	updateMic <- struct{}{}
	updateVolume <- struct{}{}
	go subscribe(updateMic, updateVolume)

	ver, err := json.Marshal(version{1})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n[\n[]\n", ver)

	gMuted, err := json.Marshal(NewGoodPanel("muted", ""))
	if err != nil {
		log.Fatal(err)
	}
	gXwayland, err := json.Marshal(NewGoodPanel("xwayland", ""))
	if err != nil {
		log.Fatal(err)
	}
	gIP, err := json.Marshal(NewGoodPanel("ip", ""))
	if err != nil {
		log.Fatal(err)
	}
	gDate, _ := json.Marshal(NewGoodPanel("date", ""))

	tXwayland := time.NewTicker(time.Minute)
	defer tXwayland.Stop()

	tTime := time.NewTicker(time.Second)
	defer tTime.Stop()

	tIP := time.NewTicker(time.Minute)
	defer tIP.Stop()

	hideVolumeDuration := 5 * time.Second
	tHideVolume := time.NewTicker(hideVolumeDuration)
	tHideVolume.Stop()
	defer tHideVolume.Stop()

	sound := Sound{}
	ip := IP{}
	date := NewDate()
	gVolume, _ := json.Marshal(volume(sound.Sink.vol, sound.Sink.ok, true))
	var diff bool

	for {
		select {
		case <-updateMic:
			gMuted, _ = json.Marshal(sound.GetSources())
		case <-updateVolume:
			diff = sound.GetSinksDiff()
			if diff {
				tHideVolume.Reset(hideVolumeDuration)
				gVolume, _ = json.Marshal(volume(sound.Sink.vol, sound.Sink.ok, false))
			}
		case <-tHideVolume.C:
			// should that be a timer not a ticker?
			tHideVolume.Stop()
			gVolume, _ = json.Marshal(volume(sound.Sink.vol, sound.Sink.ok, true))

		case <-tXwayland.C:
			gXwayland, _ = json.Marshal(GetXWayland())
		case <-tIP.C:
			gIP, _ = json.Marshal(ip.GetListeningIP())
		case <-tTime.C:
			gDate, _ = json.Marshal(date.GetDate())
		}

		fmt.Printf(",[%s,%s,%s,%s,%s]\n", gIP, gXwayland, gMuted, gVolume, gDate)
	}
}
