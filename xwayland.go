package main

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

func GetXWayland() panel {
	err := exec.Command(pidof, "Xwayland").Run()
	if err != nil {
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			ec := eerr.ExitCode()
			if ec == 1 {
				return NewGoodPanel("xwayland", "")
			}
		}
		return NewBadPanel("xwayland", "error")
	}

	return NewBadPanel("xwayland", " xwayland ")
}
