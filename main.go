package main

import (
	"flag"
	"runtime"
	"time"
)

// temp, will be removed
var (
	base   string = "#000000"
	accent string = "#000000"
)

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
	go subscribe(updateMic, updateVolume)

	bar := NewBar()
	bar.RenderInitial()

	tXwayland := time.NewTicker(time.Minute)
	defer tXwayland.Stop()

	tTime := time.NewTicker(time.Second)
	defer tTime.Stop()

	tIP := time.NewTicker(time.Minute)
	defer tIP.Stop()

	hideVolumeDuration := 5 * time.Second
	tHideVolume := time.NewTicker(hideVolumeDuration)
	defer tHideVolume.Stop()

	var diff bool

	for {
		select {
		case <-updateMic:
			bar.UpdateMuted()
		case <-updateVolume:
			diff = bar.UpdateVolume()
			if diff {
				tHideVolume.Reset(hideVolumeDuration)
			}
		case <-tHideVolume.C:
			// should that be a timer not a ticker?
			tHideVolume.Stop()
			bar.HideVolumeIfNoError()
		case <-tXwayland.C:
			bar.UpdateXWayland()
		case <-tIP.C:
			bar.UpdateIP()
		case <-tTime.C:
			bar.UpdateDate()
		}

		bar.Render()
	}
}
