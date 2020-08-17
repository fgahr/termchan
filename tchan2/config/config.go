package config

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/fgahr/termchan/tchan2"
	"github.com/pkg/errors"
)

// Opts deals with all variable and optional aspects of termchan.
type Opts struct {
	WorkingDirectory string
	Boards           map[string]tchan2.BoardMetaData
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

// New creates a new configuration object.
func New(workingDirectory string) *Opts {
	return &Opts{
		WorkingDirectory: workingDirectory,
		Boards:           nil,
	}
}

func (c *Opts) initDB(db *sql.DB) error {
	var err error

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS board (
id INTEGER PRIMARY KEY,
name TEXT UNIQUE ON CONFLICT FAIL,
description TEXT,
style TEXT,
max_threads INTEGER NOT NULL DEFAULT 50,
max_posts INTEGER NOT NULL DEFAULT 100
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

	// We initialize it here because we want to use this function for
	// refreshing a stale config as well.
	c.Boards = make(map[string]tchan2.BoardMetaData)

	boardRows, err := db.Query(`
SELECT name, descrition, style, max_threads, max_posts
FROM board;
`)
	if err != nil {
		return errors.Wrap(err, "failed to fetch board list from config file")
	}
	defer boardRows.Close()

	for boardRows.Next() {
		var md tchan2.BoardMetaData
		err = boardRows.Scan(&md.Name, &md.Description, &md.HighlightStyle, &md.MaxThreadCount, &md.MaxThreadLength)
		if err != nil {
			return errors.Wrap(err, "failed to read board definition from config file")
		}
		c.Boards[md.Name] = md
	}

	return nil
}
