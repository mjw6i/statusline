package internal

import (
	"bufio"
	"io"
)

type Bar struct {
	buf    *bufio.Writer
	date   *Date
	base   string
	accent string
	cache  struct {
		Muted    []byte
		Volume   []byte
		IP       []byte
		XWayland []byte
		Date     []byte
	}
	ip    IP
	Sound *Sound
}

func NewBar(output io.Writer, base, accent string) *Bar {
	return &Bar{
		buf:    bufio.NewWriter(output),
		Sound:  NewSound(),
		ip:     IP{},
		date:   NewDate(),
		base:   base,
		accent: accent,
	}
}

func (b *Bar) RenderInitial() {
	b.RenderHeader()
	b.UpdateAll()
	b.Render()
}

func (b *Bar) RenderHeader() {
	b.buf.WriteString(`{"version":1}`)
	b.buf.WriteString("\n[\n[]\n")
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
	ok := b.Sound.RenderMuted(&b.cache.Muted)
	b.renderPanelSuffix(&b.cache.Muted, ok)
}

func (b *Bar) UpdateVolume() bool {
	diff := b.Sound.GetSinksDiff()
	if diff {
		b.renderPanelPrefix(&b.cache.Volume, []byte("volume"))
		ok := b.Sound.RenderVolume(&b.cache.Volume, false)
		b.renderPanelSuffix(&b.cache.Volume, ok)
	}
	return diff
}

// TODO: hiding logic requires a rewrite
func (b *Bar) HideVolumeIfNoError() {
	b.renderPanelPrefix(&b.cache.Volume, []byte("volume"))
	ok := b.Sound.RenderVolume(&b.cache.Volume, true)
	b.renderPanelSuffix(&b.cache.Volume, ok)
}

func (b *Bar) UpdateIP() {
	b.renderPanelPrefix(&b.cache.IP, []byte("ip"))
	ok := b.ip.Render(&b.cache.IP)
	b.renderPanelSuffix(&b.cache.IP, ok)
}

func (b *Bar) UpdateXWayland() {
	b.renderPanelPrefix(&b.cache.XWayland, []byte("xwayland"))
	ok := RenderXWayland(&b.cache.XWayland)
	b.renderPanelSuffix(&b.cache.XWayland, ok)
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

func (b *Bar) renderPanelSuffix(buf *[]byte, ok bool) {
	*buf = append(*buf, '"')
	if !ok {
		*buf = append(*buf, []byte(`,"background":"`)...)
		*buf = append(*buf, []byte(b.accent)...)
		*buf = append(*buf, []byte(`","color":"`)...)
		*buf = append(*buf, []byte(b.base)...)
		*buf = append(*buf, '"')
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
