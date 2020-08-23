package backend

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/util"
	"github.com/pkg/errors"

	// SQLite3 bindings
	_ "github.com/mattn/go-sqlite3"
)

type sqlite struct {
	conf            *config.Opts
	boardsDirectory string
	boardDBs        map[string]*sql.DB
}

func (s *sqlite) Init() error {
	if ok, err := util.DirExists(s.boardsDirectory); err != nil {
		return err
	} else if !ok {
		err = os.Mkdir(s.boardsDirectory, 0700)
		if err != nil {
			return err
		}
	}

	for _, board := range s.conf.Boards {
		if err := s.initBoardDB(board.Name); err != nil {
			return err
		}
	}

	return nil
}

func (s *sqlite) Close() error {
	var err error
	for _, db := range s.boardDBs {
		if err == nil {
			err = db.Close()
		}
	}
	return nil
}

func (s *sqlite) PopulateBoard(boardName string, b *tchan2.BoardOverview, ok *bool) error {
	boardDB, boardOK := s.boardDBs[boardName]
	if !boardOK {
		*ok = false
		return nil
	}

	threadRows, err := boardDB.Query(`
SELECT t.topic, t.num_replies, t.created_at, t.active_at, op.author, op.content
FROM thread t INNER JOIN post op ON t.op_id = op.id
AND t.num_replies > -1 AND t.num_replies <= ?
ORDER BY t.last_reply DESC
LIMIT ?;
`)
	if err != nil {
		return errors.Wrap(err, "failed to gather thread summaries")
	}
	defer threadRows.Close()

	for threadRows.Next() {
		t := tchan2.ThreadOverview{}
		var createdTS, activeTS string
		err = threadRows.Scan(&t.Topic, &t.NumReplies, &createdTS, &activeTS,
			&t.OP.Author, &t.OP.Content)
		if err != nil {
			errors.Wrap(err, "failed to extract thread summary")
		}

		if created, err := time.Parse(time.RFC3339, createdTS); err != nil {
			return errors.Wrap(err, "malformed date string in thread table (created_at)")
		} else {
			t.OP.Timestamp = created
		}

		if active, err := time.Parse(time.RFC3339, activeTS); err != nil {
			return errors.Wrap(err, "malformed date string in thread table (active_at)")
		} else {
			t.Active = active
		}

		b.Threads = append(b.Threads, t)
	}

	return nil
}

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
    active_at TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
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

	}

	_, err = boardDB.Exec(`
CREATE TRIGGER IF NOT EXISTS update_thread_timestamp
AFTER INSERT ON post
FOR EACH ROW
BEGIN
-- SQLite3 uses UTC internally so we can add the static 'Z' time zone suffix
created_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE post.id = NEW.id;

UPDATE thread
SET num_replies = num_replies + 1,
    op_id = coalesce(op_id, NEW.id),
    created_at = coalesce(created_at, (SELECT created_at FROM post WHERE id = NEW.id)),
    active_at = max(coalesce(last_reply, '1970-01-01T00:00:00'),
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
