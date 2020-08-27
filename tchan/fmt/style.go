package fmt

import "fmt"

// Available ANSI styles
const (
	fgBlack   simpleStyle = "\u001b[30m"
	fgRed     simpleStyle = "\u001b[31m"
	fgGreen   simpleStyle = "\u001b[32m"
	fgYellow  simpleStyle = "\u001b[33m"
	fgBlue    simpleStyle = "\u001b[34m"
	fgMagenta simpleStyle = "\u001b[35m"
	fgCyan    simpleStyle = "\u001b[36m"
	fgWhite   simpleStyle = "\u001b[37m"
	ansiReset             = "\u001b[0m"
)

// Style handles formatting of a piece of output.
type Style interface {
	FormatANSI(s string) string
}

type noStyle struct{}

func (noStyle) FormatANSI(s string) string {
	return s
}

type simpleStyle string

func (sty simpleStyle) FormatANSI(s string) string {
	return fmt.Sprintf("%s%s%s", sty, s, ansiReset)
}

// GetStyle finds a formatting style by name.
func GetStyle(name string) Style {
	if s, ok := styleNames[name]; ok {
		return s
	}
	return simpleStyle(fgWhite)
}

var styleNames map[string]Style

func init() {
	styleNames = make(map[string]Style)
	styleNames["none"] = noStyle{}
	styleNames["black"] = fgBlack
	styleNames["red"] = fgRed
	styleNames["green"] = fgGreen
	styleNames["yellow"] = fgYellow
	styleNames["blue"] = fgBlue
	styleNames["magenta"] = fgMagenta
	styleNames["cyan"] = fgCyan
	styleNames["white"] = fgWhite
}
