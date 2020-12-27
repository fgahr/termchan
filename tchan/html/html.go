package html

import (
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
)

type TemplateSet struct {
	welcome *template.Template
	board   *template.Template
	thread  *template.Template
}

func (t *TemplateSet) Read(dir string) error {
	temp, err := template.ParseGlob(filepath.Join(dir, "*"))
	if err != nil {
		return errors.Wrapf(err, "failed to read templates in %s", dir)
	}

	// TODO: use some default template instead
	if welcome := temp.Lookup("welcome.template"); welcome != nil {
		t.welcome = welcome
	} else {
		return errors.Errorf("no 'welcome' template found in %s", dir)
	}

	if thread := temp.Lookup("thread.template"); thread != nil {
		t.thread = thread
	} else {
		return errors.Errorf("no 'thread' template found in %s", dir)
	}

	if board := temp.Lookup("board.template"); board != nil {
		t.board = board
	} else {
		return errors.Errorf("no 'board' template found in %s", dir)
	}

	return nil
}

type TemplateWriter struct {
	// TODO
}
