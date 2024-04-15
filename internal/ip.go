package internal

import (
	"bytes"
	"io"
	"net"
	"os/exec"
	"slices"
	"unsafe"
)

var netstat string

func init() {
	var err error
	netstat, err = exec.LookPath("netstat")
	if err != nil {
		panic(err)
	}
}

type IP struct {
	buffer bytes.Buffer
}

func processLine(line []byte) (loopback, ok bool) {
	// expecting a line like
	// tcp        0      0 127.0.0.1:44159         0.0.0.0:*               LISTEN

	var token int
	token = bytes.IndexByte(line, ' ')
	if token == -1 {
		return false, false
	}
	if !slices.Equal(line[:token], []byte("tcp")) && !slices.Equal(line[:token], []byte("tcp6")) {
		return false, false
	}
	line = line[token:]
	line = bytes.TrimSpace(line)
	line = bytes.TrimLeft(line, "0123456789")
	line = bytes.TrimSpace(line)
	line = bytes.TrimLeft(line, "0123456789")
	line = bytes.TrimSpace(line)
	token = bytes.IndexByte(line, ' ')
	if token == -1 {
		return false, false
	}
	line = line[:token]
	token = bytes.LastIndexByte(line, ':')
	if token == -1 {
		return false, false
	}
	line = line[:token]

	ipstring := unsafe.String(unsafe.SliceData(line), len(line))
	ip := net.ParseIP(ipstring)

	if ip == nil {
		return false, false
	}

	return ip.IsLoopback(), true
}

func (i *IP) GetListeningIP() (text []byte, ok bool) {
	defer i.buffer.Reset()
	res := LightCall(&i.buffer, netstat, []string{
		netstat,
		"--numeric",
		"--wide",
		"-tl",
	})
	if !res {
		return []byte("error"), false
	}

	var err error

	// skip two lines
	for range 2 {
		_, err = i.buffer.ReadBytes('\n')
		if err != nil {
			return []byte("error"), false
		}
	}

	var line []byte
	var loopback bool

	for {
		line, err = i.buffer.ReadBytes('\n')
		if len(line) > 0 {
			line = line[:len(line)-1]
			loopback, res = processLine(line)
			if !res {
				return []byte("error"), false
			}
			if !loopback {
				return []byte(" non loopback listener "), false
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return []byte("error"), false
		}
	}

	return []byte(""), true
}

func (i *IP) Render(b *[]byte) (ok bool) {
	text, ok := i.GetListeningIP()
	*b = append(*b, text...)
	return ok
}
