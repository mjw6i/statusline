package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func getListeningIP() panel {
	loopback := true
	out, err := exec.Command("netstat", "--numeric", "--wide", "-tl").Output()
	if err != nil {
		return NewBadPanel("ip", "error")
	}

	r := bytes.NewReader(out)
	s := bufio.NewScanner(r)

	var proto string
	var recv, send int64
	var local, peer string
	var ip net.IP
	format := "%s %d %d %s %s"

	skip := 2
	var index int
	for s.Scan() {
		if skip > 0 {
			skip--
			continue
		}
		// new lines could be handled by fmt, without scanner
		_, err = fmt.Sscanf(s.Text(), format, &proto, &recv, &send, &local, &peer)
		if err != nil || (proto != "tcp" && proto != "tcp6") {
			return NewBadPanel("ip", " ip error ")
		}

		index = strings.LastIndexByte(local, ':')
		if index != -1 {
			local = local[:index]
		}
		ip = net.ParseIP(local)

		if ip == nil {
			return NewBadPanel("ip", " ip error ")
		}

		if !ip.IsLoopback() {
			loopback = false
		}
	}

	if s.Err() != nil {
		return NewBadPanel("ip", " ip error ")
	}

	if !loopback {
		return NewBadPanel("ip", " non loopback listener ")
	}

	return NewGoodPanel("ip", "")
}
