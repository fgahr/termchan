package fmt

import (
	"fmt"
	"io"
)

type ansiPresenter struct {
	out io.Writer
}

func (p ansiPresenter) clean(s string) string {
	return s
}

func (p ansiPresenter) apply(sty Style, s string) string {
	return sty.FormatANSI(s)
}

func (p ansiPresenter) write(format string, args ...interface{}) error {
	if len(args) == 0 {
		_, err := fmt.Fprintln(p.out, format)
		return err
	}
	_, err := fmt.Fprintln(p.out, fmt.Sprintf(format, args...))
	return err
}

func (p ansiPresenter) header() error {
	return nil
}

func (p ansiPresenter) footer() error {
	return nil
}
