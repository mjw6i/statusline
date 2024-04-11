package internal

import (
	"time"
)

type Date struct {
	buf []byte
}

func NewDate() *Date {
	return &Date{
		buf: make([]byte, 0, 64),
	}
}

func (d *Date) getDate() {
	d.buf = d.buf[:0]
	d.buf = time.Now().AppendFormat(d.buf, "[Mon] 2006-01-02 15:04:05")
}

func (d *Date) Render(b *[]byte) (ok bool) {
	d.getDate()
	*b = append(*b, d.buf...)
	return true
}
