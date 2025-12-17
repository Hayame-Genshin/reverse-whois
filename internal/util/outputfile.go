package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type closeFunc func() error

func (c closeFunc) Close() error { return c() }

// BuildOutputWriters returns stdout writer (always) and an optional file writer.
// File behavior: overwrite by default (os.Create).
// Parent directories are NOT created; invalid paths fail with an error.
func BuildOutputWriters(stdout io.Writer, outputPath string) (io.Writer, io.Writer, io.Closer, error) {
	if outputPath == "" {
		return stdout, nil, nil, nil
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open output file %q: %w", outputPath, err)
	}

	bw := bufio.NewWriter(f)

	closer := closeFunc(func() error {
		if err := bw.Flush(); err != nil {
			_ = f.Close()
			return err
		}
		return f.Close()
	})

	return stdout, bw, closer, nil
}
