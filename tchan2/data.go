package tchan2

import (
	"time"

	"github.com/fgahr/termchan/tchan2/format/ansi"
)

// BoardMetaData contains board data that doesn't refer to its contents.
type BoardMetaData struct {
	Name          string
	Description   string
	HighlighStyle ansi.Style
}

// Post contains all data of a single post.
type Post struct {
	Author    string
	ID        int
	Timestamp time.Time
	Content   string
}

// ThreadFull contains all data of a single thread.
type ThreadFull struct {
	Board *BoardMetaData
	Topic string
	Posts []Post
}

// ThreadOverview contains superficial thread data.
type ThreadOverview struct {
	// TODO
}

// BoardOverview contains superficial board data
type BoardOverview struct {
	MetaData BoardMetaData
	Threads  []ThreadOverview
}
