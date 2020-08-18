package tchan2

import (
	"time"
)

// BoardConfig contains board data that doesn't refer to its contents.
type BoardConfig struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	HighlightStyle  string `json:"-"`
	MaxThreadCount  int    `json:"maxThreadCount"`
	MaxThreadLength int    `json:"maxThreadLength"`
}

// Post contains all data of a single post.
type Post struct {
	Author    string `json:"author"`
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

// Thread contains all data of a single thread.
type Thread struct {
	Board BoardConfig `json:"board"`
	Topic string      `json:"topic"`
	Posts []Post      `json:"posts"`
}

// ThreadOverview contains superficial thread data.
type ThreadOverview struct {
	Topic  string    `json:"topic"`
	OP     Post      `json:"op"`
	Active time.Time `json:"active"`
}

// BoardOverview contains superficial board data
type BoardOverview struct {
	MetaData BoardConfig      `json:"meta"`
	Threads  []ThreadOverview `json:"threads"`
}
