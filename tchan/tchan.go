package tchan

import (
	"time"
)

// BoardConfig contains the configured settings for a board.
type BoardConfig struct {
	Name            string `json:"name"`
	Descr           string `json:"description"`
	Style           string `json:"-"`
	MaxThreadCount  int    `json:"maxThreadCount"`
	MaxThreadLength int    `json:"maxThreadLength"`
	MaxPostBytes    int    `json:"maxPostBytes"`
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
	Board BoardConfig `json:"board"`
	Topic string      `json:"topic"`
	Posts []Post      `json:"posts"`
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
	BoardConfig                 // embedded
	Threads     []ThreadSummary `json:"threads"`
}
