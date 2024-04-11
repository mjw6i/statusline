package internal

import (
	"os/exec"
)

var pidof string

func init() {
	var err error
	pidof, err = exec.LookPath("pidof")
	if err != nil {
		panic(err)
	}
}

func GetXWayland() (text []byte, ok bool) {
	ec := LightCallExitCode(pidof, []string{pidof, "Xwayland"})
	if ec == 1 {
		return []byte(""), true
	}
	if ec == 0 {
		return []byte(" xwayland "), false
	}
	return []byte("error"), false
}

func RenderXWayland(b *[]byte) (ok bool) {
	text, ok := GetXWayland()
	*b = append(*b, text...)
	return ok
}
