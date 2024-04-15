package internal

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
)

var pactl string

func init() {
	var err error
	pactl, err = exec.LookPath("pactl")
	if err != nil {
		panic(err)
	}
}

type Sound struct {
	buffer bytes.Buffer
	Sink   struct {
		vol int
		ok  bool
	}
}

func Subscribe(ctx context.Context, buf []byte, updateSources, updateSinks chan<- struct{}) bool {
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatalln(err)
	}

	defer r.Close()

	// could use double buffering to handle reads that are not a perfect line
	go func() {
		var n int
		var line []byte
		var err error
		var i int

		for {
			line = buf
			n, err = r.Read(line)
			if err != nil {
				return
			}
			line = line[:n]
			for {
				i = bytes.IndexByte(line, '\n')
				if i == -1 {
					break
				}
				eventLine(line, updateSources, updateSinks) // not the greatest design
				line = line[i+1:]
			}

			if len(line) > 0 {
				panic("broken line")
			}
		}
	}()

	res := LightCallStreamLine(ctx, w, pactl, []string{
		pactl,
		"--format=json",
		"subscribe",
	})

	return res
}

func eventLine(line []byte, updateSources, updateSinks chan<- struct{}) {
	on, err := jsonparser.GetUnsafeString(line, "on")
	if err != nil {
		log.Fatal(err)
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

// TODO: return []byte still allocates, ignore for now
func (s *Sound) GetSources() (text []byte, ok bool) {
	res := LightCall(&s.buffer, pactl, []string{
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

	jsonparser.ArrayEach(s.buffer.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
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
		return []byte("error"), false
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
	ok := LightCall(&s.buffer, pactl, []string{
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

	jsonparser.ArrayEach(s.buffer.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
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
	text, ok := volume(s.Sink.vol, s.Sink.ok, hide)
	*b = append(*b, text...)
	return ok
}

// TODO: this chain requires a rewrite
func volume(vol int, ok bool, hide bool) (text []byte, res bool) {
	if !ok {
		return []byte("error"), false
	}

	if hide {
		return []byte(""), true
	}

	return []byte(fmt.Sprintf(" VOL: %d%% ", vol)), true
}
