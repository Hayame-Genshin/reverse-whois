package output

type Colorizer struct {
	Enabled bool
}

func NewColorizer(enabled bool) *Colorizer {
	return &Colorizer{Enabled: enabled}
}

func (c *Colorizer) Wrap(code string, s string) string {
	if c == nil || !c.Enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (c *Colorizer) Bracket(code string, content string) string {
	// Keep the brackets uncolored for a cleaner look.
	if c == nil || !c.Enabled {
		return "[" + content + "]"
	}
	return "[" + c.Wrap(code, content) + "]"
}

