package internal

import (
	"bytes"
	"context"
	"log"
	"os"
)

var NullFile *os.File

func init() {
	var err error
	NullFile, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
}

var CallEnv []string

func init() {
	CallEnv = []string{os.ExpandEnv("XDG_RUNTIME_DIR=$XDG_RUNTIME_DIR")}
}

func LightCall(buffer *bytes.Buffer, target string, args []string) bool {
	buffer.Reset()

	r, w, err := os.Pipe()
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if w != nil {
			err = w.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}
		err = r.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	files := []*os.File{NullFile, w, NullFile}
	process, err := os.StartProcess(target, args, &os.ProcAttr{
		Files: files,
		Env:   CallEnv,
		Sys:   nil,
	})
	if err != nil {
		log.Fatalln(err)
	}
	state, err := process.Wait()
	if err != nil {
		log.Fatalln(err)
	}

	err = w.Close()
	if err != nil {
		log.Fatalln(err)
	}
	w = nil

	_, err = buffer.ReadFrom(r)
	if err != nil {
		log.Fatalln(err)
	}
	// reading should most likely be done in a goroutine
	// most likely cause for this to work synchronously is linux pipe buffer
	// being bigger than output

	return state.Success()
}

// a lot of copy pasta
func LightCallStreamLine(ctx context.Context, w *os.File, target string, args []string) bool {
	defer w.Close()

	files := []*os.File{NullFile, w, NullFile}
	process, err := os.StartProcess(target, args, &os.ProcAttr{
		Files: files,
		Env:   CallEnv,
		Sys:   nil,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// could leak goroutine?
	go func() {
		<-ctx.Done()
		_ = process.Kill()
		// ignored error *perhaps* will appear in Wait
	}()

	state, err := process.Wait()
	if err != nil {
		log.Fatalln(err)
	}

	return state.Success()
}

func LightCallExitCode(target string, args []string) int {
	files := []*os.File{NullFile, NullFile, NullFile}
	process, err := os.StartProcess(target, args, &os.ProcAttr{
		Files: files,
		Env:   CallEnv,
		Sys:   nil,
	})
	if err != nil {
		log.Fatalln(err)
	}
	state, err := process.Wait()
	if err != nil {
		log.Fatalln(err)
	}

	return state.ExitCode()
}
