package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"slices"
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

func Subscribe(updateMic, updateVolume chan<- struct{}) {
	// use interruptable command to clean exit
	cmd := exec.Command(pactl, "--format=json", "subscribe")
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	eventLoop(out, updateMic, updateVolume)

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func eventLoop(r io.Reader, updateSources, updateSinks chan<- struct{}) {
	decoder := json.NewDecoder(r)

	type event struct {
		Event json.RawMessage
		On    json.RawMessage
	}

	var err error
	var e event

	for decoder.More() {
		err = decoder.Decode(&e)
		if err != nil {
			log.Fatal(err)
		}

		if slices.Equal(e.On, []byte(`"source"`)) {
			select {
			case updateSources <- struct{}{}:
			default:
			}
		} else if slices.Equal(e.On, []byte(`"sink"`)) {
			select {
			case updateSinks <- struct{}{}:
			default:
			}
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

	var muted bool
	var count int

	var class string

	jsonparser.ArrayEach(s.buffer.Bytes(), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			ok = false
			return
		}
		muted, err = jsonparser.GetBoolean(value, "mute")
		if err != nil {
			ok = false
			return
		}
		class, err = jsonparser.GetUnsafeString(value, "properties", "device.class")
		if err != nil {
			ok = false
			return
		}
		if class != "monitor" {
			count++
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
