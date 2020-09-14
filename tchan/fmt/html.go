package fmt

import (
	"fmt"
	"html"
	"io"
)

type htmlPresenter struct {
	out io.Writer
}

func (p htmlPresenter) clean(s string) string {
	return html.EscapeString(s)
}

func (p htmlPresenter) apply(sty Style, s string) string {
	return sty.FormatHTML(s)
}

func (p htmlPresenter) write(format string, args ...interface{}) error {
	if len(args) == 0 {
		_, err := fmt.Fprintln(p.out, format)
		return err
	}
	_, err := fmt.Fprintln(p.out, fmt.Sprintf(format, args...))
	return err
}

func (p htmlPresenter) header() error {
	_, err := fmt.Fprint(p.out,
		"<!doctype html>\n",
		"<html>\n",
		"<head>\n",
		"<meta charset=\"utf-8\">\n",
		"<title>termchan</title>\n",
		"<style>\n",
		"<!--\n",
		"body { color: #ffffff; background-color: #202020; }\n",
	)
	if err != nil {
		return err
	}

	for _, sty := range allStyles {
		_, err = fmt.Fprintln(p.out, sty.DefineCSS())
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(p.out,
		"-->\n",
		"</style>\n",
		"</head>\n",
		"<body>\n",
		"<pre>\n",
	)
	return err
}

func (p htmlPresenter) footer() error {
	_, err := fmt.Fprint(p.out,
		"</pre>\n",
		"</body>\n",
		"</html>\n",
	)
	return err
}
