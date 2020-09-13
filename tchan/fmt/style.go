package fmt

import "fmt"

// Style handles formatting of a piece of output.
type Style interface {
	FormatANSI(s string) string
	DefineCSS() string
	FormatHTML(s string) string
}

type noStyle struct{}

func (noStyle) FormatANSI(s string) string {
	return s
}

func (noStyle) DefineCSS() string {
	return ""
}

func (noStyle) FormatHTML(s string) string {
	return s
}

type color string

type style struct {
	name        string
	ansiNumeral int
	fg          color
}

func (sty style) FormatANSI(s string) string {
	return fmt.Sprintf("\u001b[%dm%s\u001b[0m", sty.ansiNumeral, s)
}

func (sty style) DefineCSS() string {
	return fmt.Sprintf(".%s { color: %s; }", sty.name, sty.fg)
}

func (sty style) FormatHTML(s string) string {
	return fmt.Sprintf("<span class=\"%s\">%s</span>", sty.name, s)
}

var (
	fgBlack   = style{"black", 30, "#000000"}
	fgRed     = style{"red", 31, "#ff0000"}
	fgGreen   = style{"green", 32, "#00ff00"}
	fgYellow  = style{"yellow", 33, "#ffff00"}
	fgBlue    = style{"blue", 34, "#0000ff"}
	fgMagenta = style{"magenta", 35, "#ff00ff"}
	fgCyan    = style{"cyan", 36, "#00ffff"}
	fgWhite   = style{"white", 37, "#ffffff"}
)

var allStyles []style = []style{
	{"black", 30, "#000000"},
	{"red", 31, "#ff0000"},
	{"green", 32, "#00ff00"},
	{"yellow", 33, "#ffff00"},
	{"blue", 34, "#0000ff"},
	{"magenta", 35, "#ff00ff"},
	{"cyan", 36, "#00ffff"},
	{"white", 37, "#ffffff"},
}

// getStyle finds a formatting style by name.
func getStyle(name string) Style {
	for _, sty := range allStyles {
		if sty.name == name {
			return sty
		}
	}
	return noStyle{}
}
