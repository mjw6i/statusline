package internal

import (
	"encoding/json"
	"os/exec"
)

func GetSources() panel {
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

type panel struct {
	Name                string `json:"name"`
	Text                string `json:"full_text"`
	Background          string `json:"background,omitempty"`
	Color               string `json:"color,omitempty"`
	Separator           bool   `json:"separator"`
	SeparatorBlockWidth int    `json:"separator_block_width"`
}

func NewGoodPanel(name string, text string) panel {
	return panel{Name: name, Text: text}
}

func NewBadPanel(name string, text string) panel {
	return panel{Name: name, Text: text, Color: "#000000", Background: "#000000"}
}
