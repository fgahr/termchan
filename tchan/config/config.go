package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/util"
	"github.com/pkg/errors"
)

// Settings deals with all variable and optional aspects of termchan.
type Settings struct {
	Port   int
	wd     string        `json:"-"`
	Boards []tchan.Board `json:"boards"`
}

// Defaults gives a default configuration for termchan.
func Defaults() Settings {
	return Settings{
		Port: 8088,
		wd:   "./",
		Boards: []tchan.Board{
			{
				Name:  "e",
				Descr: "example",
				Style: "red",
			},
		},
	}
}

// SetWorkingDirectory sets the working directory for termchan. Should only
// be called before initialization.
func (s *Settings) SetWorkingDirectory(dir string) error {
	if exists, err := util.DirExists(dir); err != nil {
		return errors.Wrapf(err, "failed to examine directory %s", dir)
	} else if !exists {
		return errors.Errorf("directory %s does not exist", dir)
	}
	s.wd = dir
	return nil
}

// TemplateDirectory returns the directory from where to read templates.
func (s *Settings) TemplateDirectory() string {
	return filepath.Join(s.wd, "template")
}

// BoardsDirectory returns the directory where board databases are stored.
func (s *Settings) BoardsDirectory() string {
	return filepath.Join(s.wd, "boards")
}

// ReadJSON reads settings from a JSON-encoded source.
func (s *Settings) ReadJSON(in io.Reader) error {
	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, in); err != nil {
		return err
	}

	if empty, _ := regexp.Match("^\\s*$", buf.Bytes()); empty {
		return nil
	}

	dec := json.NewDecoder(&buf)
	dec.DisallowUnknownFields()
	return dec.Decode(s)
}

// WriteJSON writes settings as JSON to a writer.
func (s *Settings) WriteJSON(out io.Writer) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "\t")
	return enc.Encode(s)
}

// ReadFromFile attempts to read a configuration file within the working
// directory.
func (s *Settings) ReadFromFile() error {
	cf := filepath.Join(s.wd, "config.json")
	if exists, err := util.FileExists(cf); err != nil {
		return errors.Wrapf(err, "error looking for config file %s", cf)
	} else if !exists {
		return nil
	}

	f, err := os.Open(cf)
	if err != nil {
		return errors.Wrapf(err, "error opening config file %s", cf)
	}
	defer f.Close()

	if err := s.ReadJSON(f); err != nil {
		return errors.Wrapf(err, "error reading config from %s", cf)
	}

	return nil
}

// BoardConfig returns the configuration for a board.
func (s *Settings) BoardConfig(boardName string) (tchan.Board, bool) {
	n := len(s.Boards)
	idx := sort.Search(n, func(i int) bool {
		return s.Boards[i].Name >= boardName
	})

	if idx == len(s.Boards) {
		return tchan.Board{}, false
	}
	b := s.Boards[idx]
	return s.Boards[idx], b.Name == boardName
}
