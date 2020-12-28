package output

import "github.com/fgahr/termchan/tchan"

// Writer describes an entity in charge of writing a server response.
type Writer interface {
	WriteWelcome(boards []tchan.BoardConfig) error
	WriteThread(thread tchan.Thread) error
	WriteBoard(board tchan.BoardOverview) error
	WriteError(status int, err error) error
}
