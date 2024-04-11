package internal

import (
	"errors"
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
	err := exec.Command(pidof, "Xwayland").Run()
	if err != nil {
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			ec := eerr.ExitCode()
			if ec == 1 {
				return []byte(""), true
			}
		}
		return []byte("error"), false
	}

	return []byte(" xwayland "), false
}

func RenderXWayland(b *[]byte) (ok bool) {
	text, ok := GetXWayland()
	*b = append(*b, text...)
	return ok
}
