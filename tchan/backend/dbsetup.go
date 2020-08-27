package backend

import (
	"database/sql"
	"path/filepath"

	"github.com/pkg/errors"
)

func (s *sqlite) initBoardDB(boardName string) error {
	path := filepath.Join(s.boardsDirectory, boardName+".db")
	var boardDB *sql.DB
	var err error

	if boardDB, err = sql.Open("sqlite3", path); err != nil {
		return errors.Wrapf(err, "failed to connect to %s", path)
	}

	_, err = boardDB.Exec(`
CREATE TABLE IF NOT EXISTS thread (
    id INTEGER PRIMARY KEY,
    op_id INTEGER,
    num_replies INTEGER DEFAULT -1,
    topic TEXT,
    created_at TEXT,
    active_at TEXT
);
`)
	if err != nil {
		return errors.Wrapf(err, "failed to create /%s/ thread table", boardName)
	}

	_, err = boardDB.Exec(`
CREATE TABLE IF NOT EXISTS post (
    id INTEGER PRIMARY KEY,
    thread_id INTEGER NOT NULL,
    author TEXT,
    author_ip TEXT,
    content TEXT NOT NULL,
    created_at TEXT,
    FOREIGN KEY(thread_id) REFERENCES thread(id) NOT DEFERRABLE
);
`)
	if err != nil {
		return errors.Wrapf(err, "failed to create /%s/ post table", boardName)
	}

	_, err = boardDB.Exec(`
CREATE INDEX IF NOT EXISTS post_by_thread_id ON post(thread_id);
`)
	if err != nil {
		return errors.Wrapf(err, "failed to create index post(thread_id) on /%s/", boardName)
	}

	_, err = boardDB.Exec(`
CREATE TRIGGER IF NOT EXISTS update_thread_timestamp
AFTER INSERT ON post
FOR EACH ROW
BEGIN
UPDATE post SET
-- SQLite3 uses UTC internally so we can add the static 'Z' time zone suffix
created_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE post.id = NEW.id;

UPDATE thread
SET num_replies = num_replies + 1,
    op_id = coalesce(op_id, NEW.id),
    created_at = coalesce(created_at, (SELECT created_at FROM post WHERE id = NEW.id)),
    -- max() doesn't work with NULL so we make sure to indeed have a value
    active_at = max(coalesce(active_at, '1970-01-01T00:00:00'),
                     (SELECT created_at FROM post WHERE id = NEW.id))
WHERE id = NEW.thread_id;
END;
`)
	if err != nil {
		return errors.Wrapf(err, "failed to create thread update trigger on /%s/", boardName)
	}

	s.boardDBs[boardName] = boardDB
	return nil
}
