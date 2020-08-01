package data

import (
	"sort"
	"time"

	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/format/ansi"
)

// Bid is the type used for Board IDs
type Bid int

// Tid is the type used for Thread IDs
type Tid int

// GPid is the type used for Global Post IDs
type GPid int

// LPid is the type used for per-board (local) Post IDs
type LPid int

// Post holds the data associated with a post.
type Post struct {
	ID        GPid      `json:"-"`
	InBoardID GPid      `json:"id"`
	Author    string    `json:"author"`
	AuthorIP  string    `json:"-"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"time"`
}

// Thread holds the data associated with a thread.
type Thread struct {
	ID    Tid            `json:"-"`
	Topic string         `json:"topic"`
	Posts []Post         `json:"posts"`
	Board *BoardOverview `json:"-"`
}

// OP returns a thread's original post to enable special formatting for it.
func (t Thread) OP() Post {
	return t.Posts[0]
}

// ThreadOverview holds all data relevant for an overview of a thread.
type ThreadOverview struct {
	Topic      string         `json:"topic"`
	OP         Post           `json:"op"`
	ThreadID   Tid            `json:"-"`
	ReplyCount int            `json:"replies"`
	Started    time.Time      `json:"postedAt"`
	LastReply  time.Time      `json:"latestReplyAt"`
	Board      *BoardOverview `json:"-"`
}

// Board holds all data associated with a board.
type Board struct {
	ID             Bid              `json:"-"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	ActiveThreads  []ThreadOverview `json:"threads"`
	HighlightStyle ansi.Style       `json:"-"`
}

type BoardOverview struct {
	ID             Bid        `json:"-"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	HighlightColor ansi.Style `json:"-"`
}

// Boards holds (cached) data about available boards.
var Boards map[string]*BoardOverview

// BoardParameters holds the globally active parameters of this instance.
type BoardParameters struct {
	Note        string           `json:"note"`
	Boards      []*BoardOverview `json:"boards"`
	PostSize    int              `json:"maxPostSize"`
	ReplyLimit  int              `json:"maxReplies"`
	ThreadLimit int              `json:"maxThreads"`
}

// GatherBoardParameters collects active global parameters.
func GatherBoardParameters() *BoardParameters {
	var overviews []*BoardOverview
	for _, board := range Boards {
		overviews = append(overviews, board)
	}
	cmp := func(i, j int) bool {
		return overviews[i].Name < overviews[j].Name
	}
	sort.Slice(overviews, cmp)

	return &BoardParameters{
		Note:        "View non-json version for detailed help",
		Boards:      overviews,
		PostSize:    config.Current.Max.PostSize,
		ReplyLimit:  config.Current.Max.PostsPerThread,
		ThreadLimit: config.Current.Max.ThreadsPerBoard,
	}
}

// BoardParams holds the active global parameters.
var BoardParams *BoardParameters
