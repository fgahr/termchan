package backend

import (
	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/config"
)

// DB handles all database interactions.
type DB interface {
	// Init sets up this database's connections.
	Init() error
	// Close destroys this database's connections.
	Close() error
	// PopulateBoard fetches a board by name.
	PopulateBoard(boardName string, b *tchan2.BoardOverview, ok *bool) error
	// PopulateThread fetches the thread with the specified post in it.
	PopulateThread(boardName string, postID int, thr *tchan2.Thread, ok *bool) error
	// CreateThread adds a new thread to a board, setting the OP's post ID.
	CreateThread(boardName string, topic string, op *tchan2.Post) error
	// AddPostToThread adds a reply to a thread, setting the post's ID in the process.
	AddReply(boardName string, postID int, post *tchan2.Post, ok *bool) error
}

// New creates a new backend which has yet to be initialized.
func New(opts *config.Opts) DB {
	return &sqlite{conf: opts}
}
