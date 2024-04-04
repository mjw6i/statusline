package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
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

	updateMic := make(chan struct{})
	updateVolume := make(chan struct{})
	go subscribe(updateMic, updateVolume)
	go func() {
		updateMic <- struct{}{}
		updateVolume <- struct{}{}
	}()

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
	gDate, _ := json.Marshal(date())
	var vol, newVol int64
	var volErr, newVolErr error
	gVolume, _ := json.Marshal(volume(vol, volErr, true))

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

	for {
		select {
		case <-updateMic:
			gMuted, _ = json.Marshal(getMics())
		case <-updateVolume:
			newVol, newVolErr = readVolume()
			if newVol != vol || newVolErr != volErr {
				vol = newVol
				volErr = newVolErr
				tHideVolume.Reset(hideVolumeDuration)
				gVolume, _ = json.Marshal(volume(vol, volErr, false))
			}
		case <-tHideVolume.C:
			// should that be a timer not a ticker?
			tHideVolume.Stop()
			gVolume, _ = json.Marshal(volume(vol, volErr, true))

		case <-tXwayland.C:
			gXwayland, _ = json.Marshal(xwayland())
		case <-tIP.C:
			gIP, _ = json.Marshal(getListeningIP())
		case <-tTime.C:
			gDate, _ = json.Marshal(date())
		}

		fmt.Printf(",[%s,%s,%s,%s,%s]\n", gIP, gXwayland, gMuted, gVolume, gDate)
	}
}

func subscribe(updateMic, updateVolume chan<- struct{}) {
	// use interruptable command to clean exit
	cmd := exec.Command("pactl", "--format=json", "subscribe")
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(out)

	type event struct {
		Event string
		On    string
	}

	for decoder.More() {
		var e event
		err = decoder.Decode(&e)
		if err != nil {
			log.Fatal(err)
		}

		switch e.On {
		case "source":
			updateMic <- struct{}{}
		case "sink":
			updateVolume <- struct{}{}
		}
	}
}

func getMics() panel {
	out, err := exec.Command("pactl", "--format=json", "list", "sources").Output()
	if err != nil {
		return NewBadPanel("mics", "error")
	}

	type source struct {
		Properties struct {
			DeviceClass string `json:"device.class"`
		}
		Mute bool
	}

	var sources []source
	err = json.Unmarshal(out, &sources)
	if err != nil {
		return NewBadPanel("mics", "error")
	}

	var count int
	var muted bool

	for _, s := range sources {
		if s.Properties.DeviceClass == "monitor" {
			continue
		}
		count++
		muted = s.Mute
	}

	if count == 0 {
		return NewGoodPanel("mics", "")
	}

	if count > 1 {
		return NewBadPanel("mics", " multiple mics ")
	}

	if muted {
		return NewGoodPanel("mics", "")
	}

	return NewBadPanel("mics", " not muted ")
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

func volume(vol int64, err error, hide bool) panel {
	if err != nil {
		return NewBadPanel("volume", "error")
	}

	if hide {
		return NewGoodPanel("volume", "")
	}

	return NewGoodPanel("volume", fmt.Sprintf(" VOL: %d%% ", vol))
}
