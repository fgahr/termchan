package backend

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/data"
	"github.com/fgahr/termchan/tchan/format/ansi"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

var backend *sql.DB

func setupBackendSchema() error {
	var err error

	_, err = backend.Exec(`
CREATE TABLE IF NOT EXISTS board (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE ON CONFLICT IGNORE,
    description TEXT NOT NULL,
    highlight_style TEXT,
    next_post_id INTEGER NOT NULL DEFAULT 1
);
`)
	if err != nil {
		return errors.Wrap(err, "failed to create 'board' table")
	}

	_, err = backend.Exec(`
CREATE TABLE IF NOT EXISTS thread (
    id INTEGER PRIMARY KEY,
    board_id INTEGER,
    op_id INTEGER,
    num_replies INTEGER DEFAULT -1,
    topic TEXT,
    created_at TEXT,
    last_reply TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);

CREATE INDEX IF NOT EXISTS thread_by_board ON thread(board_id);
CREATE INDEX IF NOT EXISTS thread_by_op ON thread(op_id);
CREATE INDEX IF NOT EXISTS thread_by_update ON thread(last_reply);
`)
	if err != nil {
		return errors.Wrap(err, "failed to create 'thread' table")
	}

	_, err = backend.Exec(`
CREATE TABLE IF NOT EXISTS post (
    id INTEGER PRIMARY KEY,
    thread_id INTEGER NOT NULL,
    in_board_id INTEGER,
    author TEXT NOT NULL DEFAULT 'Anonymous',
    author_ip TEXT,
    content TEXT NOT NULL,
    created_at TEXT,
    FOREIGN KEY(thread_id) REFERENCES thread(id) NOT DEFERRABLE
);
`)
	if err != nil {
		return errors.Wrap(err, "failed to create 'post' table")
	}

	_, err = backend.Exec(`
CREATE INDEX IF NOT EXISTS post_by_thread_id ON post(thread_id);
`)
	if err != nil {
		return errors.Wrap(err, "failed to create index on 'post' table")
	}

	_, err = backend.Exec(`
CREATE TRIGGER IF NOT EXISTS increment_in_board_id
AFTER INSERT ON post
FOR EACH ROW
BEGIN
-- Update post table
UPDATE post SET
in_board_id = (SELECT b.next_post_id FROM board b
               WHERE b.id = (SELECT t.board_id FROM thread t WHERE t.id = NEW.thread_id)),
-- SQLite3 uses UTC internally so we can add the static 'Z' time zone suffix
created_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE post.id = NEW.id;
-- Update board table
UPDATE board SET next_post_id = next_post_id + 1
WHERE id = (SELECT thread.board_id FROM thread WHERE thread.id = NEW.thread_id);
-- Update thread table
UPDATE thread
SET num_replies = num_replies + 1,
    op_id = coalesce(op_id, NEW.id),
    created_at = coalesce(created_at, (SELECT created_at FROM post WHERE id = NEW.id)),
    last_reply = max(coalesce(last_reply, '1970-01-01T00:00:00'),
                     (SELECT created_at FROM post WHERE id = NEW.id))
WHERE id = NEW.thread_id;
END;
`)

	return nil
}

func createDefaultBoards() error {
	for _, b := range config.Current.Boards {
		_, err := backend.Exec(`
INSERT INTO board(name, description, highlight_style)
SELECT ?, ?, ?
WHERE NOT EXISTS(SELECT 1 FROM board WHERE name = ?);
`, b.Name, b.Desc, b.HiLi, b.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// FetchBoardIDs inserts known boards into a board overview map.
func FetchBoardIDs() (map[string]*data.BoardOverview, error) {
	boards := make(map[string]*data.BoardOverview)
	var err error

	rows, err := backend.Query("SELECT id, name, description, highlight_style FROM board;")
	if err != nil {
		return boards, errors.Wrap(err, "failed to fetch table info from the database")
	}
	defer rows.Close()

	for rows.Next() {
		to := data.BoardOverview{}
		var colorName string
		if err = rows.Scan(&to.ID, &to.Name, &to.Description, &colorName); err != nil {
			return boards, errors.Wrap(err, "error fetching table IDs")
		}
		to.HighlightColor = ansi.GetStyle(colorName)
		boards[to.Name] = &to
	}

	return boards, nil
}

// GetThreadID finds the db-side thread ID from a pair of board and in-board ID.
func GetThreadID(boardID data.Bid, inBoardPostID data.LPid) (data.Tid, bool, error) {
	rows, err := backend.Query(`
SELECT t.id
FROM thread t
INNER JOIN post p
ON p.id = t.op_id
WHERE t.board_id = ?
AND p.in_board_id = ?
LIMIT 1;
`, boardID, inBoardPostID)
	if err != nil {
		log.Println("error encountered", err)
		return 0, false, errors.Wrapf(err, "failed to retrieve thread ID for /%d/%d", boardID, inBoardPostID)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, false, nil
	}
	var id data.Tid
	if err = rows.Scan(&id); err != nil {
		return 0, false, errors.Wrapf(err, "failed to parse thread ID for /%d/%d", boardID, inBoardPostID)
	}
	return id, true, nil
}

// GetBoard fetches the data required to print a board overview.
func GetBoard(boardID data.Bid) (*data.Board, error) {
	board := data.Board{ID: boardID}
	board.ActiveThreads = []data.ThreadOverview{}
	var err error

	boardRows, err := backend.Query(`
SELECT name, description, highlight_style FROM board WHERE id = ?;
`, boardID)
	if err != nil {
		return &board, errors.Wrap(err, "failed to fetch board")
	}

	if !boardRows.Next() {
		return &board, errors.Wrap(err, "no such board")
	}

	var styleName string
	if err = boardRows.Scan(&board.Name, &board.Description, &styleName); err != nil {
		boardRows.Close()
		return &board, errors.Wrap(err, "error evaluating database query")
	}
	boardRows.Close()
	board.HighlightStyle = ansi.GetStyle(styleName)

	threadRows, err := backend.Query(`
SELECT t.id, t.topic, t.num_replies, t.created_at, t.last_reply,
    op.in_board_id, op.author, op.content
FROM thread t INNER JOIN post op ON t.op_id = op.id
WHERE t.board_id = ?
AND t.num_replies > -1 AND t.num_replies <= ?
ORDER BY t.last_reply DESC
LIMIT ?;
`, boardID, config.Current.Max.PostsPerThread, config.Current.Max.ThreadsPerBoard)
	if err != nil {
		return &board, errors.Wrap(err, "failed to gather thread summaries")
	}
	defer threadRows.Close()

	for threadRows.Next() {
		t := data.ThreadOverview{}
		var createdTS, replyTS string
		err = threadRows.Scan(&t.ThreadID, &t.Topic, &t.ReplyCount, &createdTS, &replyTS,
			&t.OP.InBoardID, &t.OP.Author, &t.OP.Content)
		if err != nil {
			return &board, errors.Wrap(err, "failed to extract thread summary")
		}
		if created, err := time.Parse(time.RFC3339, createdTS); err != nil {
			return &board, errors.Wrap(err, "malformed date string in thread table (created_at)")
		} else {
			t.Started = created
			t.OP.Timestamp = created
		}

		if latest, err := time.Parse(time.RFC3339, replyTS); err != nil {
			return &board, errors.Wrap(err, "malformed date string in thread table (latest_reply)")
		} else {
			t.LastReply = latest
		}

		t.Board = data.Boards[board.Name]
		board.ActiveThreads = append(board.ActiveThreads, t)
	}

	return &board, nil
}

// GetThread fetches all data required for a thread overview.
func GetThread(threadID data.Tid) (*data.Thread, error) {
	var err error
	var thread data.Thread

	threadRows, err := backend.Query("SELECT id, topic FROM thread WHERE id = ?;", threadID)
	if err != nil {
		return &thread, errors.Wrapf(err, "couldn't find thread with ID %d", threadID)
	}
	defer threadRows.Close()

	if !threadRows.Next() {
		return &thread, errors.Errorf("no thread with ID: %d", threadID)
	}
	threadRows.Scan(&thread.ID, &thread.Topic)
	threadRows.Close()

	postRows, err := backend.Query(`
SELECT id, in_board_id, author, content, created_at FROM post
WHERE post.thread_id = ?
ORDER BY post.created_at ASC;`, threadID)
	if err != nil {
		return &thread, errors.Wrapf(err, "unable to find posts for thread (id=%d)", threadID)
	}
	defer postRows.Close()
	for postRows.Next() {
		var post data.Post
		ts := ""
		err = postRows.Scan(&post.ID, &post.InBoardID, &post.Author, &post.Content, &ts)
		if err != nil {
			log.Println("unable to extract post data from database row")
			continue
		}
		post.Timestamp, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			log.Printf("invalid timestamp in post table (id=%d): %s\n", post.ID, ts)
			continue
		}

		thread.Posts = append(thread.Posts, post)
	}

	return &thread, err
}

// CreateThread creates a new thread on the selected board with the given topic.
func CreateThread(boardID data.Bid, topic string) (data.Tid, error) {
	result, err := backend.Exec("INSERT INTO thread (board_id, topic) VALUES (?, ?);", boardID, topic)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to create thread on board %d with topic '%s'", boardID, topic)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "failed to retrieve last insert id")
	}
	return data.Tid(id), nil
}

// AddReplyToThread adds a reply to an existing thread.
func AddReplyToThread(threadID data.Tid, post *data.Post) error {
	result, err := backend.Exec(
		"INSERT INTO post (thread_id, author, author_ip, content) VALUES (?, ?, ?, ?);",
		threadID, post.Author, post.AuthorIP, post.Content)
	if err != nil {
		return errors.Wrapf(err, "failed to persist post {%v}", post)
	} else if n, _ := result.RowsAffected(); n == 0 {
		return errors.Errorf("failed to persist post {%v}", post)
	} else if pid, err := result.LastInsertId(); err == nil {
		post.ID = data.GPid(pid)
	} else {
		return errors.New("failed to determine post id")
	}

	return nil
}

// Connect sets up a database connection.
func Connect() error {
	dbFile := config.Current.DBFile
	var err error

	if _, err = os.Stat(dbFile); os.IsNotExist(err) {
		_, err = os.Create(dbFile)
	} else if err != nil {
		return errors.Wrapf(err, "unable to locate or create database file: %s", dbFile)
	}

	if backend, err = sql.Open("sqlite3", dbFile); err != nil {
		return errors.Wrapf(err, "unable to open database file: %s", dbFile)
	}

	if err = setupBackendSchema(); err != nil {
		return errors.Wrap(err, "failed to setup database schema")
	}

	if err = createDefaultBoards(); err != nil {
		return errors.Wrap(err, "failed to populate board table")
	}

	return nil
}

// Shutdown closes the database connection.
func Shutdown() error {
	return backend.Close()
}
