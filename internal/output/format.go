package output

import "strings"

// Output format:
// [SEARCH_TERM] [MODE] [SEARCH_TYPE] [key: value]
func FormatKeyValueLine(c *Colorizer, term, mode, searchType, key, value string) string {
	if c == nil {
		c = NewColorizer(false)
	}

	seg1 := c.Bracket("36", term)       // cyan
	seg2 := c.Bracket("35", mode)       // magenta
	seg3 := c.Bracket("34", searchType) // blue
	seg4 := c.Bracket("33", key+": "+value)

	return strings.Join([]string{seg1, seg2, seg3, seg4}, " ")
}

func FormatErrorLine(c *Colorizer, term, mode, searchType, key, msg string) string {
	if c == nil {
		c = NewColorizer(false)
	}

	seg1 := c.Bracket("36", term)
	seg2 := c.Bracket("35", mode)
	seg3 := c.Bracket("34", searchType)
	seg4 := c.Bracket("31", key+": "+msg) // red

	return strings.Join([]string{seg1, seg2, seg3, seg4}, " ")
}
