package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type Bar struct {
	muted    []byte
	volume   []byte
	IP       []byte // naming is beyond bad
	XWayland []byte
	Date     []byte // terrible

	buf *bufio.Writer

	sound Sound
	ip    IP
	date  *Date
}

func NewBar(output io.Writer) *Bar {
	return &Bar{
		buf:   bufio.NewWriter(output),
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
	enc := json.NewEncoder(b.buf)
	err := enc.Encode(version{1})
	if err != nil {
		log.Fatal(err)
	}
	b.buf.WriteString("[\n[]\n")
	b.buf.Flush()
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
	// this could be done cleaner
	b.buf.WriteByte(',')
	b.buf.WriteByte('[')
	b.buf.Write(b.IP)
	b.buf.WriteByte(',')
	b.buf.Write(b.XWayland)
	b.buf.WriteByte(',')
	b.buf.Write(b.muted)
	b.buf.WriteByte(',')
	b.buf.Write(b.volume)
	b.buf.WriteByte(',')
	b.buf.Write(b.Date)
	b.buf.WriteByte(']')
	b.buf.WriteByte('\n')
	b.buf.Flush()
}
