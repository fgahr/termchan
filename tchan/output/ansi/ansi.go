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
	post    *template.Template
	board   *template.Template
	thread  *template.Template
}

func placeholders() template.FuncMap {
	nothing := func(v interface{}) string { return "" }
	return template.FuncMap{
		"formatBoard": nothing,
		"formatPost":  nothing,
		"highlight":   nothing,
		"timeANSIC":   nothing,
	}
}

func (t *TemplateSet) UseDefaults() {
	// Needs to be replaced before template execution
	t.welcome = template.Must(
		template.New("welcome.template").
			Funcs(placeholders()).
			Parse(output.DefaultWelcome))
	t.post = template.Must(
		template.New("post.template").
			Funcs(placeholders()).
			Parse(output.DefaultPost))
	t.board = template.Must(
		template.New("board.template").
			Funcs(placeholders()).
			Parse(output.DefaultBoard))
	t.thread = template.Must(
		template.New("thread.template").
			Funcs(placeholders()).
			Parse(output.DefaultThread))
}

func (t *TemplateSet) Read(wd string) error {
	// reset to defaults, then look for alternatives
	t.UseDefaults()
	// FIXME: find error handling concept; folder may not exist etc.
	tdir := filepath.Join(wd, "template")
	temp, err := template.ParseGlob(filepath.Join(tdir, "*.template"))
	if err != nil {
		return errors.Wrapf(err, "failed to read templates in %s", tdir)
	}

	if welcome := temp.Lookup("welcome.template"); welcome != nil {
		t.welcome = welcome
	}

	if post := temp.Lookup("post.template"); post != nil {
		t.post = post
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
		Funcs(template.FuncMap{
			"formatBoard": func(b tchan.BoardConfig) string {
				return w.formatBoard(b)
			}}).
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

func (w *Writer) formatBoard(board tchan.BoardConfig) string {
	styles := map[string]string{
		"black":   "\u001b[30m",
		"red":     "\u001b[31m",
		"green":   "\u001b[32m",
		"yellow":  "\u001b[33m",
		"blue":    "\u001b[34m",
		"magenta": "\u001b[35m",
		"cyan":    "\u001b[36m",
		"white":   "\u001b[37m",
	}

	if style, ok := styles[board.Style]; ok {
		return fmt.Sprintf(
			"/%s%s\u001b[0m/ - %s%s\u001b[0m",
			style, board.Name, style, board.Descr)
	}
	return fmt.Sprintf("/%s/ - %s", board.Name, board.Descr)
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
