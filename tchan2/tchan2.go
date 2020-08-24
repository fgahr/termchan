package tchan2

import (
	"time"
)

// BoardConfig contains the configured settings for a board.
type BoardConfig struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	HighlightStyle  string `json:"-"`
	MaxThreadCount  int    `json:"maxThreadCount"`
	MaxThreadLength int    `json:"maxThreadLength"`
	MaxPostBytes    int    `json:"maxPostBytes"`
}

// Post contains all data of a single post.
type Post struct {
	ID        int       `json:"id"`
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

// ThreadOverview contains superficial thread data.
type ThreadOverview struct {
	Topic      string    `json:"topic"`
	OP         Post      `json:"op"`
	NumReplies int       `json:"numReplies"`
	Active     time.Time `json:"active"`
}

// BoardOverview contains superficial board data.
type BoardOverview struct {
	MetaData BoardConfig      `json:"meta"`
	Threads  []ThreadOverview `json:"threads"`
}
