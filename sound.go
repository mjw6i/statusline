package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

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

func getSources() panel {
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

func getSinks() (int, error) {
	var flp, frp int
	out, err := exec.Command("pactl", "--format=json", "list", "sinks").Output()
	if err != nil {
		return 0, err
	}

	type sink struct {
		Volume struct {
			FrontLeft struct {
				P string `json:"value_percent"`
			} `json:"front-left"`
			FrontRight struct {
				P string `json:"value_percent"`
			} `json:"front-right"`
		}
		Mute bool
	}

	var sinks []sink
	err = json.Unmarshal(out, &sinks)
	if err != nil {
		return 0, err
	}

	if len(sinks) != 1 {
		return 0, errors.New("expected one sink")
	}

	flp, err = strconv.Atoi(strings.TrimSuffix(sinks[0].Volume.FrontLeft.P, "%"))
	if err != nil {
		return 0, err
	}
	frp, err = strconv.Atoi(strings.TrimSuffix(sinks[0].Volume.FrontRight.P, "%"))
	if err != nil {
		return 0, err
	}

	if flp != frp {
		return 0, errors.New("uneven channels")
	}

	return flp, nil
}

func volume(vol int, err error, hide bool) panel {
	if err != nil {
		return NewBadPanel("volume", "error")
	}

	if hide {
		return NewGoodPanel("volume", "")
	}

	return NewGoodPanel("volume", fmt.Sprintf(" VOL: %d%% ", vol))
}
