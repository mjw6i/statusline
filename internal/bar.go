package internal

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type Bar struct {
	buf   *bufio.Writer
	date  *Date
	cache struct {
		Muted    []byte
		Volume   []byte
		IP       []byte
		XWayland []byte
		Date     []byte
	}
	ip    IP
	sound Sound
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
	b.renderPanelPrefix(&b.cache.Muted, []byte("muted"))
	ok := b.sound.RenderMuted(&b.cache.Muted)
	b.renderPanelSuffix(&b.cache.Muted, ok)
}

func (b *Bar) UpdateVolume() bool {
	diff := b.sound.GetSinksDiff()
	if diff {
		b.renderPanelPrefix(&b.cache.Volume, []byte("volume"))
		ok := b.sound.RenderVolume(&b.cache.Volume, false)
		b.renderPanelSuffix(&b.cache.Volume, ok)
	}
	return diff
}

// TODO: hiding logic requires a rewrite
func (b *Bar) HideVolumeIfNoError() {
	b.renderPanelPrefix(&b.cache.Volume, []byte("volume"))
	ok := b.sound.RenderVolume(&b.cache.Volume, true)
	b.renderPanelSuffix(&b.cache.Volume, ok)
}

func (b *Bar) UpdateIP() {
	var err error
	b.cache.IP, err = json.Marshal(b.ip.GetListeningIP())
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bar) UpdateXWayland() {
	var err error
	b.cache.XWayland, err = json.Marshal(GetXWayland())
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Bar) UpdateDate() {
	b.renderPanelPrefix(&b.cache.Date, []byte("date"))
	ok := b.date.Render(&b.cache.Date)
	b.renderPanelSuffix(&b.cache.Date, ok)
}

// TODO: less appends, perhaps use different function
func (b *Bar) renderPanelPrefix(buf *[]byte, name []byte) {
	*buf = (*buf)[:0]
	*buf = append(*buf, '{')
	*buf = append(*buf, []byte(`"separator":false,`)...)
	*buf = append(*buf, []byte(`"separator_block_width":0,`)...)
	*buf = append(*buf, []byte(`"name":"`)...)
	*buf = append(*buf, name...)
	*buf = append(*buf, []byte(`","full_text":`)...)
	*buf = append(*buf, '"')
}

// TODO: pass color values
func (b *Bar) renderPanelSuffix(buf *[]byte, ok bool) {
	*buf = append(*buf, '"')
	if !ok {
		*buf = append(*buf, []byte(`,"background":"#000000"`)...)
		*buf = append(*buf, []byte(`,"color":"#000000"`)...)
	}
	*buf = append(*buf, '}')
}

func (b *Bar) Render() {
	// this could be done cleaner
	b.buf.WriteByte(',')
	b.buf.WriteByte('[')
	b.buf.Write(b.cache.IP)
	b.buf.WriteByte(',')
	b.buf.Write(b.cache.XWayland)
	b.buf.WriteByte(',')
	b.buf.Write(b.cache.Muted)
	b.buf.WriteByte(',')
	b.buf.Write(b.cache.Volume)
	b.buf.WriteByte(',')
	b.buf.Write(b.cache.Date)
	b.buf.WriteByte(']')
	b.buf.WriteByte('\n')
	b.buf.Flush()
}
