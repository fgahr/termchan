package backend

import (
	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan2"
)

// DB handles all database interactions.
type DB interface {
	// Init sets up this database's connections.
	Init() error
	// BoardExists return whether a board with the name exists.
	BoardExists(boardName string) bool
	// ThreadExists queries for the existence of a thread by name and post ID.
	ThreadExists(boardName string, postID int) bool
	// GetBoard fetches a board by name.
	GetBoard(boardName string) (tchan2.BoardOverview, error)
	// PopulateBoard fetches a board by name.
	PopulateBoard(boardName string, b *tchan2.BoardOverview) error
	// GetThread fetches the thread with the specified post in it.
	GetThread(boardName string, postID int) (tchan2.Thread, error)
	// PopulateThread fetches the thread with the specified post in it.
	PopulateThread(boardName string, postID int, thr *tchan2.Thread) error
	// CreateThread adds a new thread to a board, setting the OP's post ID.
	CreateThread(boardName string, topic string, op *tchan2.Post) error
	// AddPostToThread adds a post to a thread, setting the post's ID.
	AddAsReply(boardName string, postID int, post *tchan2.Post) error
}

// NewDB creates a new backend which has yet to be initialized.
func NewDB(opts *config.Opts) DB {
	// TODO
	return nil
}
