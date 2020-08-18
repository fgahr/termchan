package fmt

import (
	"fmt"
	"io"

	"github.com/fgahr/termchan/tchan2"
)

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

// Writer describes an entity in charge of writing a server response.
type Writer interface {
	WriteWelcome(boardData []tchan2.BoardConfig) error
	WriteThread(thread tchan2.ThreadFull) error
	WriteBoard(board tchan2.BoardOverview) error
	WriteError(err error) error
}

// GetWriter finds a named writer as expected for format specifications.
func GetWriter(name string, w io.Writer) Writer {
	return newJSONWriter(w)
}

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
	if s, ok := styleNames[name]; ok {
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
