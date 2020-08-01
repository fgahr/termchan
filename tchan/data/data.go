package data

import (
	"sort"
	"time"

	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/format/ansi"
)

// Board ID type
type Bid int

// Thread ID type
type Tid int

// Global Post ID type
type GPid int

// Per-board (local) Post ID type
type LPid int

type Post struct {
	ID        GPid      `json:"-"`
	InBoardId GPid      `json:"id"`
	Author    string    `json:"author"`
	AuthorIP  string    `json:"-"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"time"`
}

type Thread struct {
	ID    Tid            `json:"-"`
	Topic string         `json:"topic"`
	Posts []Post         `json:"posts"`
	Board *BoardOverview `json:"-"`
}

func (t Thread) OP() Post {
	return t.Posts[0]
}

type ThreadOverview struct {
	Topic      string         `json:"topic"`
	OP         Post           `json:"op"`
	ThreadID   Tid            `json:"-"`
	ReplyCount int            `json:"replies"`
	Started    time.Time      `json:"postedAt"`
	LastReply  time.Time      `json:"latestReplyAt"`
	Board      *BoardOverview `json:"-"`
}

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

var Boards map[string]*BoardOverview

type HelpContent struct {
	Note        string           `json:"note"`
	Boards      []*BoardOverview `json:"boards"`
	PostSize    int              `json:"maxPostSize"`
	ReplyLimit  int              `json:"maxReplies"`
	ThreadLimit int              `json:"maxThreads"`
}

func GatherHelpContent() *HelpContent {
	var overviews []*BoardOverview
	for _, board := range Boards {
		overviews = append(overviews, board)
	}
	cmp := func(i, j int) bool {
		return overviews[i].Name < overviews[j].Name
	}
	sort.Slice(overviews, cmp)

	return &HelpContent{
		Note:        "View non-json version for detailed help",
		Boards:      overviews,
		PostSize:    config.Conf.Max.PostSize,
		ReplyLimit:  config.Conf.Max.PostsPerThread,
		ThreadLimit: config.Conf.Max.ThreadsPerBoard,
	}
}

var Help *HelpContent
