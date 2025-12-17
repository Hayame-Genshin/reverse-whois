package output

import (
	"fmt"
	"io"
	"sync"
)

type Printer struct {
	w      io.Writer
	colors *Colorizer
	mu     sync.Mutex
}

func NewPrinter(w io.Writer, c *Colorizer) *Printer {
	return &Printer{w: w, colors: c}
}

func (p *Printer) Colors() *Colorizer {
	return p.colors
}

func (p *Printer) Println(line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, _ = fmt.Fprintln(p.w, line)
}
