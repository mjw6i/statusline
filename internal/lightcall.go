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
	// this only works synchronously due to linux pipe buffer size
	// being bigger than output

	return state.Success()
}

type StreamLineFunction func(line []byte)

// a lot of copy pasta
func LightCallStreamLine(ctx context.Context, buf []byte, target string, args []string, fun StreamLineFunction) bool {
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
				fun(line[:i])
				line = line[i+1:]
			}

			if len(line) > 0 {
				panic("broken line")
			}
		}
	}()

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

	err = w.Close()
	if err != nil {
		log.Fatalln(err)
	}
	w = nil

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
