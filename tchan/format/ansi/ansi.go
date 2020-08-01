package ansi

const (
	FgBlack   = "\u001b[30m"
	FgRed     = "\u001b[31m"
	FgGreen   = "\u001b[32m"
	FgYellow  = "\u001b[33m"
	FgBlue    = "\u001b[34m"
	FgMagenta = "\u001b[35m"
	FgCyan    = "\u001b[36m"
	FgWhite   = "\u001b[37m"
	Reset     = "\u001b[0m"
)

// Style is an ANSI formatting string
type Style string

// GetStyle finds a pre-defined ANSI formatting string by name.
func GetStyle(name string) Style {
	s, ok := styleNames[name]
	if ok {
		return s
	}
	return FgWhite
}

var styleNames map[string]Style

func init() {
	styleNames = make(map[string]Style)
	styleNames["black"] = FgBlack
	styleNames["red"] = FgRed
	styleNames["green"] = FgGreen
	styleNames["yellow"] = FgYellow
	styleNames["blue"] = FgBlue
	styleNames["magenta"] = FgMagenta
	styleNames["cyan"] = FgCyan
	styleNames["white"] = FgWhite
}
