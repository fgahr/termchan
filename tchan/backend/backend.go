package backend

import (
	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/config"
)

// DB handles all database interactions.
type DB interface {
	// Init sets up this database's connections.
	Init() error

	// Close destroys this database's connections.
	Close() error

	// Refresh renews this database's connections.
	Refresh() error

	// PopulateBoard fetches a board by name.
	PopulateBoard(boardName string, b *tchan.BoardOverview, ok *bool) error

	// PopulateThread fetches the thread with the specified post in it.
	PopulateThread(boardName string, postID int64, thr *tchan.Thread, ok *bool) error

	// CreateThread adds a new thread to a board, setting the OP's post ID.
	CreateThread(boardName string, topic string, op *tchan.Post) error

	// AddPostToThread adds a reply to a thread, setting the post's ID in the process.
	AddReply(boardName string, postID int64, post *tchan.Post, ok *bool) error
}

// New creates a new backend which has yet to be initialized.
func New(opts *config.Opts) DB {
	return &sqlite{conf: opts}
}
