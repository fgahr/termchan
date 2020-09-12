package config

import (
	"testing"

	"github.com/fgahr/termchan/tchan"
)

func confWithBoards(boards ...string) Opts {
	c := Opts{}
	for _, b := range boards {
		c.Boards = append(c.Boards, tchan.BoardConfig{Name: b})
	}
	return c
}

func TestBoardExists(t *testing.T) {
	c := confWithBoards("a", "b", "d")
	if _, ok := c.BoardConfig("b"); !ok {
		t.Errorf("expected /b/ to exist but it didn't")
	}
	if _, ok := c.BoardConfig("c"); ok {
		t.Errorf("expected /c/ to not exist but it did")
	}
}
