package config

import "fmt"

// Board holds configuration options for a single board.
type Board struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	HiLi string `json:"-"`
}

// Limits holds global limits for this service.
type Limits struct {
	ThreadsPerBoard int `json:"threadsPerBoard"`
	PostsPerThread  int `json:"postsPerThread"`
	PostSize        int `json:"postSize"`
}

// Opts holds a set of config options.
type Opts struct {
	Port   int
	DBFile string
	Max    Limits
	Boards []Board
}

// Conf is the globally accessible current configuration.
var Conf *Opts

// PortString gives the local address to be used in the web server setup.
func (c *Opts) PortString() string {
	return fmt.Sprintf(":%d", c.Port)
}

func newDefault() *Opts {
	return &Opts{
		Port:   8088,
		DBFile: "file.db",
		Max: Limits{
			ThreadsPerBoard: 50,
			PostsPerThread:  128,
			PostSize:        8192,
		},
		Boards: []Board{
			Board{"b", "Random", "yellow"},
			Board{"g", "Technology", "green"},
			Board{"v", "Games", "cyan"},
			Board{"m", "Meta", "blue"},
		},
	}
}

func init() {
	Conf = newDefault()
}
