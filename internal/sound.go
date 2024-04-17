package internal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
)

const SUBSCRIBE_BUFFER_LENGTH = 128

var pactl string

func init() {
	var err error
	pactl, err = exec.LookPath("pactl")
	if err != nil {
		panic(err)
	}
}

type Sound struct {
	subBuffer []byte
	cmdBuffer bytes.Buffer
	Sink      struct {
		vol int
		ok  bool
	}
}

func NewSound() *Sound {
	buf := make([]byte, SUBSCRIBE_BUFFER_LENGTH)
	return &Sound{subBuffer: buf}
}

func (s *Sound) Subscribe(ctx context.Context, updateSources, updateSinks chan<- struct{}) bool {
	return subscribe(ctx, s.subBuffer, updateSources, updateSinks)
}

func subscribe(ctx context.Context, buf []byte, updateSources, updateSinks chan<- struct{}) bool {
	p := func(r *os.File) {
		objectCallback(buf, r, func(b []byte) {
			eventLine(b, updateSources, updateSinks)
		})
	}

	res := LightCallStream(ctx, p, pactl, []string{
		pactl,
		"--format=json",
		"subscribe",
	})

	return res
}

// buf should be a len=cap scratchpad
func objectCallback(buf []byte, r *os.File, cb func([]byte)) {
	var n int
	var err error
	var filled int

	for {
		if cap(buf) == filled {
			fmt.Printf("Space %d Filled %d Diff %d\n", cap(buf), filled, cap(buf)-filled)
			panic("buffer out of space") // could extend the buffer instead
		}
		n, err = r.Read(buf[filled:]) // append data
		filled += n
		if err != nil {
			return
		}
		for {
			object, dataType, offset, err := jsonparser.Get(buf[:filled])
			if dataType != jsonparser.Object || err != nil {
				break
			}
			cb(object)
			// buf = [<?delim><event1><delim><event2><delim><part of event3>]
			// remove to the end of event1, let parser handle new lines before/after
			copy(buf, buf[offset:])
			filled -= offset
		}
	}
}

func eventLine(object []byte, updateSources, updateSinks chan<- struct{}) {
	on, err := jsonparser.GetUnsafeString(object, "on")
	if err != nil {
		panic(err)
	}

	switch on {
	case "source":
		select {
		case updateSources <- struct{}{}:
		default:
		}
	case "sink":
		select {
		case updateSinks <- struct{}{}:
		default:
		}
	}
}

func (s *Sound) GetSources() (text []byte, ok bool) {
	res := LightCall(&s.cmdBuffer, pactl, []string{
		pactl,
		"--format=json",
		"list",
		"sources",
	})
	if !res {
		return []byte("error"), false
	}

	ok = true

	var count int
	var muted bool

	var devClass string
	var devMuted bool

	jsonparser.ArrayEach(s.cmdBuffer.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			ok = false
			return
		}
		devMuted, err = jsonparser.GetBoolean(value, "mute")
		if err != nil {
			ok = false
			return
		}
		devClass, err = jsonparser.GetUnsafeString(value, "properties", "device.class")
		if err != nil {
			ok = false
			return
		}
		if devClass != "monitor" {
			count++
			muted = devMuted
		}
	})

	if !ok {
		return []byte(" error "), false
	}

	if count == 0 {
		return []byte(""), true
	}

	if count > 1 {
		return []byte(" multiple mics "), false
	}

	if muted {
		return []byte(""), true
	}

	return []byte(" not muted "), false
}

func (s *Sound) RenderMuted(b *[]byte) (ok bool) {
	text, ok := s.GetSources()
	*b = append(*b, text...)
	return ok
}

func (s *Sound) GetSinksDiff() (diff bool) {
	vol, ok := s.GetSinks()
	if vol == s.Sink.vol && s.Sink.ok == ok {
		return false
	}
	s.Sink.vol = vol
	s.Sink.ok = ok
	return true
}

func (s *Sound) GetSinks() (int, bool) {
	ok := LightCall(&s.cmdBuffer, pactl, []string{
		pactl,
		"--format=json",
		"list",
		"sinks",
	})

	if !ok {
		return 0, false
	}

	var flp, frp string
	var count int

	jsonparser.ArrayEach(s.cmdBuffer.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			ok = false
			return
		}

		flp, err = jsonparser.GetUnsafeString(value, "volume", "front-left", "value_percent")
		if err != nil {
			ok = false
			return
		}

		frp, err = jsonparser.GetUnsafeString(value, "volume", "front-right", "value_percent")
		if err != nil {
			ok = false
			return
		}
		count++
	})

	if !ok || count != 1 {
		return 0, false
	}

	var fl, fr int
	var err error

	fl, err = strconv.Atoi(strings.TrimSuffix(flp, "%"))
	if err != nil {
		return 0, false
	}
	fr, err = strconv.Atoi(strings.TrimSuffix(frp, "%"))
	if err != nil {
		return 0, false
	}

	if fl != fr {
		return 0, false
	}

	return fl, true
}

func (s *Sound) RenderVolume(b *[]byte, hide bool) (ok bool) {
	ok = s.Sink.ok
	if !ok {
		*b = append(*b, []byte(" error ")...)
	} else if !hide {
		*b = append(*b, []byte(" VOL: ")...)
		*b = append(*b, []byte(strconv.Itoa(s.Sink.vol))...)
		*b = append(*b, []byte("% ")...)
	}
	return ok
}
