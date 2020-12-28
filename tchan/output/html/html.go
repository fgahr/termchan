package html

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

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

func (w *Writer) withHeaderAndFooter(f func() error) error {
	if _, err := w.out.Write([]byte(header)); err != nil {
		return errors.Wrap(err, "writing HTML header failed")
	}

	if err := f(); err != nil {
		return err
	}

	if _, err := w.out.Write([]byte(footer)); err != nil {
		return errors.Wrap(err, "writing HTML footer failed")
	}

	return nil
}

func formatBoard(board tchan.BoardConfig) template.HTML {
	return template.HTML(fmt.Sprintf(
		"/<span class=%q>%s</span>/ - <span class=%q>%s</span>",
		board.Style, board.Name, board.Style, board.Descr))
}

func (w *Writer) WriteWelcome(boards []tchan.BoardConfig) error {
	return w.withHeaderAndFooter(func() error {
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
	})
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
	FgBlack   template.HTML
	FgRed     template.HTML
	FgGreen   template.HTML
	FgYellow  template.HTML
	FgBlue    template.HTML
	FgMagenta template.HTML
	FgCyan    template.HTML
	FgWhite   template.HTML
	End       template.HTML
	Separator struct {
		Single template.HTML
		Double template.HTML
	}
}

var defaults Defaults = Defaults{
	FgBlack:   "<span class=\"black\">",
	FgRed:     "<span class=\"red\">",
	FgGreen:   "<span class=\"green\">",
	FgYellow:  "<span class=\"yellow\">",
	FgBlue:    "<span class=\"blue\">",
	FgMagenta: "<span class=\"magenta\">",
	FgCyan:    "<span class=\"cyan\">",
	FgWhite:   "<span class=\"white\">",
	End:       "</span>",
	Separator: struct {
		Single template.HTML
		Double template.HTML
	}{
		Single: "<span class=\"black\">--------------------------------------------------------------------------------</span>",
		Double: "<span class=\"black\">================================================================================</span>",
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
