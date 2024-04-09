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
	tHideVolume.Stop()
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

type Bar struct {
	muted    []byte
	volume   []byte
	IP       []byte // naming is beyond bad
	XWayland []byte
	Date     []byte // terrible

	sound Sound
	ip    IP
	date  *Date
}

func NewBar() *Bar {
	return &Bar{
		sound: Sound{},
		ip:    IP{},
		date:  NewDate(),
	}
}

func (b *Bar) RenderInitial() {
	b.RenderHeader()
	b.UpdateAll()
	b.Render()
}

func (b *Bar) RenderHeader() {
	ver, err := json.Marshal(version{1})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n[\n[]\n", ver)
}

func (b *Bar) UpdateAll() {
	b.UpdateMuted()
	b.UpdateVolume()
	b.UpdateIP()
	b.UpdateXWayland()
	b.UpdateDate()
}

func (b *Bar) UpdateMuted() {
	var err error
	b.muted, err = json.Marshal(b.sound.GetSources())
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bar) UpdateVolume() bool {
	diff := b.sound.GetSinksDiff()
	var err error
	if diff {
		b.volume, err = json.Marshal(volume(b.sound.Sink.vol, b.sound.Sink.ok, false))
	}
	if err != nil {
		log.Fatal(err)
	}
	return diff
}

func (b *Bar) HideVolumeIfNoError() {
	var err error
	b.volume, err = json.Marshal(volume(b.sound.Sink.vol, b.sound.Sink.ok, true))
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bar) UpdateIP() {
	var err error
	b.IP, err = json.Marshal(b.ip.GetListeningIP())
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bar) UpdateXWayland() {
	var err error
	b.XWayland, err = json.Marshal(GetXWayland())
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bar) UpdateDate() {
	var err error
	b.Date, err = json.Marshal(b.date.GetDate())
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bar) Render() {
	fmt.Printf(",[%s,%s,%s,%s,%s]\n", b.IP, b.XWayland, b.muted, b.volume, b.Date)
}
