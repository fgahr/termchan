package tchan

import (
	"time"
)

const (
	maxThreadsDefault      = 50
	maxThreadLengthDefault = 100
	maxPostBytesDefault    = 4096
)

// Board contains the configured settings for a board.
type Board struct {
	Name            string `json:"name"`
	Descr           string `json:"description"`
	Style           string `json:"style"`
	ThreadsMax      int    `json:"maxThreads,omitempty"`
	ThreadLengthMax int    `json:"maxThreadLength,omitempty"`
	PostBytesMax    int    `json:"maxPostBytes,omitempty"`
}

// MaxThreads returns the maximum number of active threads to be displayed on
// this board.
func (b Board) MaxThreads() int {
	if b.ThreadsMax > 0 {
		return b.ThreadsMax
	}
	return maxThreadsDefault
}

// MaxThreadLength returns the maximum number of posts a thread can have and
// be considered active.
func (b Board) MaxThreadLength() int {
	if b.ThreadLengthMax > 0 {
		return b.ThreadLengthMax
	}
	return maxThreadsDefault
}

// MaxPostBytes returns the maximum length (in bytes) for post content on this
// board.
func (b Board) MaxPostBytes() int {
	if b.PostBytesMax > 0 {
		return b.PostBytesMax
	}
	return maxPostBytesDefault
}

// Post contains all data of a single post.
type Post struct {
	ID        int64     `json:"id"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
}

// Thread contains all data of a single thread.
type Thread struct {
	Board Board  `json:"board"`
	Topic string `json:"topic"`
	Posts []Post `json:"posts"`
}

// ID returns the thread's associated ID, i.e. the OP's post ID.
func (t Thread) ID() int64 {
	return t.Posts[0].ID
}

// NumReplies returns the number of replies this thread has received.
func (t Thread) NumReplies() int {
	return len(t.Posts) - 1
}

// ThreadSummary contains superficial thread data.
type ThreadSummary struct {
	Topic      string    `json:"topic"`
	OP         Post      `json:"op"`
	NumReplies int       `json:"numReplies"`
	Active     time.Time `json:"active"`
}

// ID returns the thread's associated ID, i.e. the OP's post ID.
func (t ThreadSummary) ID() int64 {
	return t.OP.ID
}

// BoardOverview contains superficial board data.
type BoardOverview struct {
	Board                   // embedded
	Threads []ThreadSummary `json:"threads"`
}
