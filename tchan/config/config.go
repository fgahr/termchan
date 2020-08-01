package config

import "fmt"

type Board struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	HiLi string `json:"-"`
}

type Limits struct {
	ThreadsPerBoard int `json:"threadsPerBoard"`
	PostsPerThread  int `json:"postsPerThread"`
	PostSize        int `json:"postSize"`
}

type Config struct {
	Port    int     `json:"port"`
	DBFile  string  `json:"dbFile"`
	LogFile string  `json:"logFile"`
	Max     Limits  `json:"limits"`
	Boards  []Board `json:"boards"`
}

var Conf *Config

func (c *Config) PortString() string {
	return fmt.Sprintf(":%d", c.Port)
}

func newDefault() *Config {
	return &Config{
		Port:    8088,
		DBFile:  "file.db",
		LogFile: "",
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
