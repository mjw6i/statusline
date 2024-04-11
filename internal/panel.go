package internal

import (
	"encoding/json"
)

type panel struct {
	Name                string          `json:"name"`
	Background          string          `json:"background,omitempty"`
	Color               string          `json:"color,omitempty"`
	Text                json.RawMessage `json:"full_text"`
	SeparatorBlockWidth int             `json:"separator_block_width"`
	Separator           bool            `json:"separator"`
}

func NewGoodPanel(name string, text string) panel {
	// temp
	b := make([]byte, 0, len(text)+2)
	b = append(b, '"')
	b = append(b, []byte(text)...)
	b = append(b, '"')
	return NewGoodPanelFast(name, b)
}

func NewGoodPanelFast(name string, text []byte) panel {
	return panel{Name: name, Text: text}
}

func NewBadPanel(name string, text string) panel {
	// temp
	b := make([]byte, 0, len(text)+2)
	b = append(b, '"')
	b = append(b, []byte(text)...)
	b = append(b, '"')
	return NewBadPanelFast(name, b)
}

func NewBadPanelFast(name string, text []byte) panel {
	// return panel{Name: name, Text: text, Color: base, Background: accent}
	return panel{Name: name, Text: text, Color: "#000000", Background: "#000000"}
}
