package output

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/util"
	"github.com/pkg/errors"
)

// Writer describes an entity in charge of writing a server response.
type Writer interface {
	WriteWelcome(boards []tchan.Board) error
	WriteThread(thread tchan.Thread) error
	WriteBoard(board tchan.BoardOverview) error
	WriteError(status int, err error) error
}

func writeTemplate(tdir string, fname string, content []byte) error {
	path := filepath.Join(tdir, fname)

	if exists, err := util.FileExists(path); err != nil {
		return err
	} else if !exists {
		if err := ioutil.WriteFile(path, content, 0644); err != nil {
			return errors.Wrapf(err, "failed to write template")
		}
	}

	return nil
}

// WriteTemplates dumps the default templates inside the given directory.
func WriteTemplates(dir string) error {
	if exists, err := util.DirExists(dir); err != nil {
		return errors.Wrapf(err, "unable to check out template directory %s", dir)
	} else if !exists {
		if err := os.Mkdir(dir, 0755); err != nil {
			return errors.Wrapf(err, "unable to create template directory %s", dir)
		}
	}

	if err := writeTemplate(dir, "welcome.template", []byte(DefaultWelcome)); err != nil {
		return err
	}

	if err := writeTemplate(dir, "post.template", []byte(DefaultPost)); err != nil {
		return err
	}

	if err := writeTemplate(dir, "thread.template", []byte(DefaultThread)); err != nil {
		return err
	}

	if err := writeTemplate(dir, "board.template", []byte(DefaultBoard)); err != nil {
		return err
	}

	if err := writeTemplate(dir, "error.template", []byte(DefaultError)); err != nil {
		return err
	}

	return nil
}
