package util

import (
	"fmt"
	"io"
)

type Logger struct {
	w       io.Writer
	enabled bool
	silent  bool
}

func NewLogger(w io.Writer, enabled bool, silent bool) *Logger {
	return &Logger{w: w, enabled: enabled, silent: silent}
}

func (l *Logger) Debugf(format string, args ...any) {
	if l == nil || l.silent || !l.enabled {
		return
	}
	_, _ = fmt.Fprintf(l.w, "[debug] "+format+"\n", args...)
}
