package format

import "fmt"

// Available ANSI styles
const (
	fgBlack   = "\u001b[30m"
	fgRed     = "\u001b[31m"
	fgGreen   = "\u001b[32m"
	fgYellow  = "\u001b[33m"
	fgBlue    = "\u001b[34m"
	fgMagenta = "\u001b[35m"
	fgCyan    = "\u001b[36m"
	fgWhite   = "\u001b[37m"
	ansiReset = "\u001b[0m"
)

// Style handles formatting of a piece of output.
type Style interface {
	FormatANSI(s string) string
}

type style string

func (sty style) FormatANSI(s string) string {
	return fmt.Sprintf("%s%s%s", sty, s, ansiReset)
}

// GetStyle finds a formatting style by name.
func GetStyle(name string) Style {
	s, ok := styleNames[name]
	if ok {
		return s
	}
	return style(fgWhite)
}

var styleNames map[string]style

func init() {
	styleNames = make(map[string]style)
	styleNames["black"] = fgBlack
	styleNames["red"] = fgRed
	styleNames["green"] = fgGreen
	styleNames["yellow"] = fgYellow
	styleNames["blue"] = fgBlue
	styleNames["magenta"] = fgMagenta
	styleNames["cyan"] = fgCyan
	styleNames["white"] = fgWhite
}
