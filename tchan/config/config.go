package config

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"

	"github.com/fgahr/termchan/tchan"
	"github.com/pkg/errors"
)

// Opts deals with all variable and optional aspects of termchan.
type Opts struct {
	WorkingDirectory string
	Boards           []tchan.BoardConfig
}

// New creates a new configuration object.
func New(workingDirectory string) *Opts {
	return &Opts{
		WorkingDirectory: workingDirectory,
		Boards:           nil,
	}
}

func (c *Opts) connectDB() (*sql.DB, error) {
	var err error

	dbFile := filepath.Join(c.WorkingDirectory, "global.db")
	if _, err = os.Stat(dbFile); os.IsNotExist(err) {
		var f *os.File
		f, err = os.Create(dbFile)
		f.Close()
	}

	if err != nil {
		return nil, errors.Wrapf(err, "unable to locate or create database file: %s", dbFile)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open sqlite file: %s", dbFile)
	}

	return db, nil
}

func (c *Opts) initDB(db *sql.DB) error {
	var err error

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS board (
id INTEGER PRIMARY KEY,
name TEXT NOT NULL ON CONFLICT FAIL UNIQUE ON CONFLICT FAIL,
description TEXT NOT NULL,
style TEXT NOT NULL DEFAULT 'none',
max_threads INTEGER NOT NULL DEFAULT 50,
max_posts INTEGER NOT NULL DEFAULT 100,
max_post_bytes INTEGER NOT NULL DEFAULT 4096
);
`)

	return errors.Wrap(err, "unable to create board configuration table")
}

func (c *Opts) Read() error {
	db, err := c.connectDB()
	if err != nil {
		return errors.Wrap(err, "unable to open config file")
	}
	defer db.Close()

	if err = c.initDB(db); err != nil {
		return err
	}

	// We initialize it here because we want to use this function for
	// refreshing a stale config as well.
	c.Boards = make([]tchan.BoardConfig, 0)

	boardRows, err := db.Query(`
SELECT name, description, style, max_threads, max_posts, max_post_bytes
FROM board
ORDER BY name ASC;
`)
	if err != nil {
		return errors.Wrap(err, "failed to fetch board list from config file")
	}
	defer boardRows.Close()

	for boardRows.Next() {
		var bc tchan.BoardConfig
		err = boardRows.Scan(&bc.Name, &bc.Description, &bc.HighlightStyle,
			&bc.MaxThreadCount, &bc.MaxThreadLength, &bc.MaxPostBytes)
		if err != nil {
			return errors.Wrap(err, "failed to read board definition from config file")
		}
		c.Boards = append(c.Boards, bc)
	}

	return nil
}

// BoardConfig returns the configuration for a board.
func (c *Opts) BoardConfig(boardName string) (tchan.BoardConfig, bool) {
	n := len(c.Boards)
	idx := sort.Search(n, func(i int) bool {
		return c.Boards[i].Name >= boardName
	})

	if idx == len(c.Boards) {
		return tchan.BoardConfig{}, false
	}
	b := c.Boards[idx]
	return c.Boards[idx], b.Name == boardName
}
