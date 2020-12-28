package ansi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/output"
	"github.com/fgahr/termchan/tchan/util"
	"github.com/pkg/errors"
)

type TemplateSet struct {
	welcome *template.Template
	post    *template.Template
	board   *template.Template
	thread  *template.Template
	error   *template.Template
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
	t.error = template.Must(
		template.New("error.template").
			Funcs(placeholders()).
			Parse(output.DefaultError))
}

func parseTemplateFile(name string, dir string) (*template.Template, error) {
	path := filepath.Join(dir, name)
	if exists, err := util.FileExists(path); err != nil {
		return nil, errors.Wrapf(err, "unable to check out template file %s", path)
	} else if !exists {
		// Nothing to do here
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open template file %s", path)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read from template file %s", path)
	}

	tmpl, err := template.New(name).Funcs(placeholders()).Parse(string(content))
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing template file %s", path)
	}

	return tmpl, nil
}

func (t *TemplateSet) Read(wd string) error {
	// reset to defaults, then look for alternatives
	t.UseDefaults()
	tdir := filepath.Join(wd, "template")
	if exists, err := util.DirExists(tdir); err != nil {
		return errors.Wrapf(err, "unable to check out template directory %s", tdir)
	} else if !exists {
		// Nothing to do, just use the defaults
		return nil
	}

	if tmpl, err := parseTemplateFile("welcome.template", tdir); err != nil {
		return err
	} else if tmpl != nil {
		t.welcome = tmpl
	}

	if tmpl, err := parseTemplateFile("post.template", tdir); err != nil {
		return err
	} else if tmpl != nil {
		t.post = tmpl
	}

	if tmpl, err := parseTemplateFile("thread.template", tdir); err != nil {
		return err
	} else if tmpl != nil {
		t.thread = tmpl
	}

	if tmpl, err := parseTemplateFile("board.template", tdir); err != nil {
		return err
	} else if tmpl != nil {
		t.thread = tmpl
	}

	if tmpl, err := parseTemplateFile("error.template", tdir); err != nil {
		return err
	} else if tmpl != nil {
		t.error = tmpl
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
			"formatBoard": w.boardFormatter(),
		}).
		Execute(w.out, payload)
}

func (w *Writer) WriteThread(thread tchan.Thread) error {
	payload := struct {
		Defaults     // embedded
		tchan.Thread // embedded
	}{
		Defaults: defaults,
		Thread:   thread,
	}
	return w.temp.thread.
		Funcs(template.FuncMap{
			"formatPost":  w.postFormatter(thread.Board.Style),
			"formatBoard": w.boardFormatter(),
			"highlight":   w.highlighter(thread.Board.Style),
			"timeANSIC":   w.timeFormatter(time.ANSIC),
		}).
		Execute(w.out, payload)
}

func (w *Writer) WriteBoard(board tchan.BoardOverview) error {
	payload := struct {
		Defaults            // embedded
		tchan.BoardOverview // embedded
	}{
		Defaults:      defaults,
		BoardOverview: board,
	}

	return w.temp.board.
		Funcs(template.FuncMap{
			"formatPost":  w.postFormatter(board.Style),
			"formatBoard": w.boardFormatter(),
			"highlight":   w.highlighter(board.Style),
			"timeANSIC":   w.timeFormatter(time.ANSIC),
		}).
		Execute(w.out, payload)
}

func (w *Writer) WriteError(status int, err error) error {
	w.out.WriteHeader(status)
	payload := struct {
		Defaults // embedded
		Status   int
		Error    string
	}{
		Defaults: defaults,
		Status:   status,
		Error:    err.Error(),
	}

	return w.temp.error.
		Funcs(template.FuncMap{
			"timeANSIC": w.timeFormatter(time.ANSIC),
			"highlight": w.highlighter("red"),
		}).
		Execute(w.out, payload)
}

func (w *Writer) postFormatter(styleName string) func(tchan.Post) string {
	return func(p tchan.Post) string {
		payload := struct {
			Defaults   // embedded
			tchan.Post // embedded
		}{
			Defaults: defaults,
			Post:     p,
		}
		buf := bytes.Buffer{}
		err := w.temp.post.Funcs(template.FuncMap{
			"highlight": w.highlighter(styleName),
			"timeANSIC": w.timeFormatter(time.ANSIC),
		}).Execute(&buf, payload)
		if err != nil {
			return ""
		}
		return buf.String()
	}
}

func (w *Writer) boardFormatter() func(tchan.BoardConfig) string {
	return func(b tchan.BoardConfig) string {
		return w.formatBoard(b)
	}
}

func (w *Writer) highlighter(styleName string) func(string) string {
	return func(s string) string {
		if sty, ok := w.style(styleName); ok {
			return fmt.Sprintf("%s%s\u001b[0m", sty, s)
		}
		return s
	}
}

func (w *Writer) timeFormatter(format string) func(time.Time) string {
	return func(t time.Time) string {
		return t.Format(format)
	}
}

func (w *Writer) formatBoard(board tchan.BoardConfig) string {
	if style, ok := w.style(board.Style); ok {
		return fmt.Sprintf(
			"/%s%s\u001b[0m/ - %s%s\u001b[0m",
			style, board.Name, style, board.Descr)
	}
	// TODO: log warning?
	return fmt.Sprintf("/%s/ - %s", board.Name, board.Descr)
}

func (w *Writer) style(name string) (string, bool) {
	switch name {
	case "black":
		return "\u001b[30m", true
	case "red":
		return "\u001b[31m", true
	case "green":
		return "\u001b[32m", true
	case "yellow":
		return "\u001b[33m", true
	case "blue":
		return "\u001b[34m", true
	case "magenta":
		return "\u001b[35m", true
	case "cyan":
		return "\u001b[36m", true
	case "white":
		return "\u001b[37m", true
	default:
		return "\u001b[37m", false
	}
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
