package logging

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"
)

// Init configures the standard logger to use aligned columns.
func Init() {
	log.SetFlags(0)
	log.SetOutput(&columnWriter{out: os.Stderr})
}

type columnWriter struct {
	out io.Writer
}

func (w *columnWriter) Write(p []byte) (int, error) {
	lines := bytes.Split(p, []byte("\n"))
	written := 0
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		stamp := time.Now().Format("2006-01-02 15:04:05")
		prefix := []byte("INFO\t| " + stamp + "\t | ")
		n, err := w.out.Write(append(prefix, append(line, '\n')...))
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}
