package internal

import (
	"bytes"
	"io"
	"net"
	"os/exec"
	"slices"
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
	ip := net.ParseIP(string(line))
	if ip == nil {
		return false, false
	}

	return ip.IsLoopback(), true
}

func (i *IP) GetListeningIP() panel {
	defer i.buffer.Reset()
	cmd := exec.Command(netstat, "--numeric", "--wide", "-tl")
	cmd.Stdout = &i.buffer
	err := cmd.Run()
	if err != nil {
		return NewBadPanel("ip", "error")
	}

	// skip two lines
	for range 2 {
		_, err = i.buffer.ReadBytes('\n')
		if err != nil {
			return NewBadPanel("ip", "error")
		}
	}

	var line []byte
	var loopback, ok bool

	for {
		line, err = i.buffer.ReadBytes('\n')
		if len(line) > 0 {
			line = line[:len(line)-1]
			loopback, ok = processLine(line)
			if !ok {
				return NewBadPanel("ip", "error")
			}
			if !loopback {
				return NewBadPanel("ip", " non loopback listener ")
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return NewBadPanel("ip", "error")
		}
	}

	return NewGoodPanel("ip", "")
}
