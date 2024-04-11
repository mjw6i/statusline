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

func (d *Date) GetDate() panel {
	d.buf = d.buf[:0]
	d.buf = append(d.buf, '"')
	d.buf = time.Now().AppendFormat(d.buf, "[Mon] 2006-01-02 15:04:05")
	d.buf = append(d.buf, '"')

	return NewGoodPanelFast("date", d.buf)
}
