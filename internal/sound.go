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

	type source struct {
		Properties struct {
			DeviceClass string `json:"device.class"`
		}
		Mute bool
	}

	var sources []source
	err := json.Unmarshal(s.buffer.Bytes(), &sources)
	if err != nil {
		return []byte("error"), false
	}

	var count int
	var muted bool

	for _, s := range sources {
		if s.Properties.DeviceClass == "monitor" {
			continue
		}
		count++
		muted = s.Mute
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

	var flp, frp int
	if !ok {
		return 0, false
	}

	type sink struct {
		Volume struct {
			FrontLeft struct {
				P string `json:"value_percent"`
			} `json:"front-left"`
			FrontRight struct {
				P string `json:"value_percent"`
			} `json:"front-right"`
		}
		Mute bool
	}

	var sinks []sink
	err := json.Unmarshal(s.buffer.Bytes(), &sinks)
	if err != nil {
		return 0, false
	}

	if len(sinks) != 1 {
		return 0, false
	}

	flp, err = strconv.Atoi(strings.TrimSuffix(sinks[0].Volume.FrontLeft.P, "%"))
	if err != nil {
		return 0, false
	}
	frp, err = strconv.Atoi(strings.TrimSuffix(sinks[0].Volume.FrontRight.P, "%"))
	if err != nil {
		return 0, false
	}

	if flp != frp {
		return 0, false
	}

	return flp, true
}

func volume(vol int, ok bool, hide bool) panel {
	if !ok {
		return NewBadPanel("volume", "error")
	}

	if hide {
		return NewGoodPanel("volume", "")
	}

	return NewGoodPanel("volume", fmt.Sprintf(" VOL: %d%% ", vol))
}
