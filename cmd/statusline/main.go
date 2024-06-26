package main

import (
	"context"
	"flag"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/mjw6i/statusline/internal"
)

func init() {
	// debug.SetGCPercent(20)
	runtime.GOMAXPROCS(1)
}

func main() {
	var base, accent string
	flag.StringVar(&base, "base", "#000000", "base color")
	flag.StringVar(&accent, "accent", "#000000", "accent color")
	flag.Parse()

	bar := internal.NewBar(os.Stdout, base, accent)

	updateMic := make(chan struct{}, 1)
	updateVolume := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ok := bar.Sound.Subscribe(ctx, updateMic, updateVolume)
		log.Fatalln(ok)
	}()

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
