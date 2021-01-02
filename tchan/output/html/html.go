package html

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

// UseDefaults resets the template set to its defaults.
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

// Read attempts to read recognized templates from within the given directory.
// Missing template files will be substituted with the defaults.
func (t *TemplateSet) Read(dir string) error {
	// Reset to defaults, then look for alternatives
	t.UseDefaults()
	if exists, err := util.DirExists(dir); err != nil {
		return errors.Wrapf(err, "unable to check out template directory %s", dir)
	} else if !exists {
		// Nothing to do, just use the defaults
		return nil
	}

	if tmpl, err := parseTemplateFile("welcome.template", dir); err != nil {
		return err
	} else if tmpl != nil {
		t.welcome = tmpl
	}

	if tmpl, err := parseTemplateFile("post.template", dir); err != nil {
		return err
	} else if tmpl != nil {
		t.post = tmpl
	}

	if tmpl, err := parseTemplateFile("thread.template", dir); err != nil {
		return err
	} else if tmpl != nil {
		t.thread = tmpl
	}

	if tmpl, err := parseTemplateFile("board.template", dir); err != nil {
		return err
	} else if tmpl != nil {
		t.board = tmpl
	}

	if tmpl, err := parseTemplateFile("error.template", dir); err != nil {
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

func formatBoard(board tchan.Board) template.HTML {
	return template.HTML(fmt.Sprintf(
		"/<span class=%q>%s</span>/ - <span class=%q>%s</span>",
		board.Style, board.Name, board.Style, board.Descr))
}

func (w *Writer) WriteWelcome(boards []tchan.Board) error {
	return w.withHeaderAndFooter(func() error {
		payload := struct {
			Defaults // embedded
			Boards   []tchan.Board
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
	return w.withHeaderAndFooter(func() error {
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
	})
}

func (w *Writer) WriteBoard(board tchan.BoardOverview) error {
	return w.withHeaderAndFooter(func() error {
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
	})
}

func (w *Writer) WriteError(status int, err error) error {
	return w.withHeaderAndFooter(func() error {
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
	})
}

func (w *Writer) postFormatter(styleName string) func(tchan.Post) template.HTML {
	return func(p tchan.Post) template.HTML {
		p.Content = html.EscapeString(p.Content)
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
			log.Println(err)
			return ""
		}
		return template.HTML(buf.String())
	}
}

func (w *Writer) boardFormatter() func(tchan.Board) template.HTML {
	return func(b tchan.Board) template.HTML {
		return w.formatBoard(b)
	}
}

func (w *Writer) highlighter(styleName string) func(interface{}) template.HTML {
	return func(v interface{}) template.HTML {
		return template.HTML(fmt.Sprintf("<span class=%q>%v</span>", styleName, v))
	}
}

func (w *Writer) timeFormatter(format string) func(time.Time) template.HTML {
	return func(t time.Time) template.HTML {
		return template.HTML(t.Format(format))
	}
}

func (w *Writer) formatBoard(board tchan.Board) template.HTML {
	return template.HTML(fmt.Sprintf(
		"/<span class=%q>%s</span>/ - <span class=%q>%s</span>",
		board.Style, board.Name, board.Style, board.Descr))
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
<meta charset="utf-8">
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
