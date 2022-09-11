package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/util"
)

// Settings deals with all variable and optional aspects of termchan.
type Settings struct {
	Transport Transport     `json:"transport"`
	wd        string        `json:"-"`
	Boards    []tchan.Board `json:"boards"`
}

type Protocol int

const (
	TCP Protocol = iota
	Unix
)

func (p Protocol) String() string {
	switch p {
	case TCP:
		return "tcp"
	case Unix:
		return "unix"
	default:
		return ""
	}
}

func (p Protocol) MarshalJSON() ([]byte, error) {
	switch p {
	case TCP:
		return json.Marshal("tcp")
	case Unix:
		return json.Marshal("unix")
	default:
		return nil, errors.Errorf("unknown protocol value: %d", int(p))
	}
}

func (p *Protocol) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case "tcp":
		*p = TCP
	case "unix":
		*p = Unix
	default:
		return errors.Errorf("invalid protocol: %s", s)
	}

	return nil
}

type Transport struct {
	Protocol Protocol `json:"protocol"`
	Socket   string   `json:"socket"`
}

func (t Transport) String() string {
	switch t.Protocol {
	case TCP:
		return "tcp" + t.Socket
	case Unix:
		return "unix:" + t.Socket
	default:
		return "unknown " + t.Socket
	}
}

// Defaults gives a default configuration for termchan.
func Defaults() Settings {
	return Settings{
		Transport: Transport{
			Protocol: TCP,
			Socket:   ":8088",
		},
		wd: "./",
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
		// No error, just use defaults.
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
