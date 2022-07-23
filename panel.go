package main

type version struct {
	Version int `json:"version"`
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
	return panel{Name: name, Text: text, Color: base, Background: accent}
}
