package ansi

import (
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/output"
	"github.com/pkg/errors"
)

type TemplateSet struct {
	welcome *template.Template
	board   *template.Template
	thread  *template.Template
}

func (t *TemplateSet) UseDefaults() {
	t.welcome = template.Must(
		template.New("welcome.template").
			Funcs(template.FuncMap{"formatBoard": formatBoard}).
			Parse(output.DefaultWelcome))
	// TODO
	t.board = nil
	t.thread = nil
}

func (t *TemplateSet) Read(wd string) error {
	// reset to defaults, then look for alternatives
	t.UseDefaults()
	// FIXME: find error handling concept; folder may not exist etc.
	tdir := filepath.Join(wd, "template")
	temp, err := template.ParseGlob(filepath.Join(tdir, "*"))
	if err != nil {
		return errors.Wrapf(err, "failed to read templates in %s", tdir)
	}

	if welcome := temp.Lookup("welcome.template"); welcome != nil {
		t.welcome = welcome
	}

	if thread := temp.Lookup("thread.template"); thread != nil {
		t.thread = thread
	}

	if board := temp.Lookup("board.template"); board != nil {
		t.board = board
	}

	return nil
}

type Writer struct {
	req  *http.Request
	out  http.ResponseWriter
	temp TemplateSet
}

func NewWriter(r *http.Request, w http.ResponseWriter, ts TemplateSet) *Writer {
	return &Writer{req: r, out: w, temp: ts}
}

func formatBoard(board tchan.BoardConfig) string {
	// FIXME: lookup styles by name
	return fmt.Sprintf(
		"/<span class=%q>%s</span>/ - <span class=%q>%s</span>",
		board.HighlightStyle, board.Name, board.HighlightStyle, board.Description)
}

func (w *Writer) WriteWelcome(boards []tchan.BoardConfig) error {
	payload := struct {
		Defaults // embedded
		Boards   []tchan.BoardConfig
		Hostname string
	}{
		Defaults: defaults,
		Boards:   boards,
		Hostname: w.req.Host,
	}

	return w.temp.welcome.
		Funcs(template.FuncMap{"formatBoard": formatBoard}).
		Execute(w.out, payload)
}

func (w *Writer) WriteThread(thread tchan.Thread) error {
	// TODO
	return nil
}

func (w *Writer) WriteBoard(board tchan.BoardOverview) error {
	// TODO
	return nil
}

func (w *Writer) WriteError(status int, err error) error {
	// TODO
	return nil
}

type Defaults struct {
	FgBlack   string
	FgRed     string
	FgGreen   string
	FgYellow  string
	FgBlue    string
	FgMagenta string
	FgCyan    string
	FgWhite   string
	End       string
	Separator struct {
		Single string
		Double string
	}
}

var defaults Defaults = Defaults{
	FgBlack:   "\u001b[30m",
	FgRed:     "\u001b[31m",
	FgGreen:   "\u001b[32m",
	FgYellow:  "\u001b[33m",
	FgBlue:    "\u001b[34m",
	FgMagenta: "\u001b[35m",
	FgCyan:    "\u001b[36m",
	FgWhite:   "\u001b[37m",
	End:       "\u001b[0m",
	Separator: struct {
		Single string
		Double string
	}{
		Single: "\u001b[30m--------------------------------------------------------------------------------\u001b[0m",
		Double: "\u001b[30m================================================================================\u001b[0m",
	},
}

const header = `<!doctype html>
<html>
<head>
<meta charset=\"utf-8\">
<title>termchan</title>
<style>
<!--
body { color: #ffffff; background-color: #202020; }
.black { color: #000000; }
.red { color: #ff0000; }
.green { color: #00ff00; }
.yellow { color: #ffff00; }
.blue { color: #0000ff; }
.magenta { color: #ff00ff; }
.cyan { color: #00ffff; }
.white { color: #ffffff; }
-->
</style>
</head>
<body>
<pre>
`

const footer = `</pre>
</body>
</html>
`
