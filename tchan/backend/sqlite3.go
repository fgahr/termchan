package backend

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/util"
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
	s.boardsDirectory = filepath.Join(s.conf.WorkingDirectory, "boards")
	if ok, err := util.DirExists(s.boardsDirectory); err != nil {
		return err
	} else if !ok {
		err = os.Mkdir(s.boardsDirectory, 0700)
		if err != nil {
			return err
		}
	}

	boards := make(map[string]*sql.DB)
	for _, board := range s.conf.Boards {
		bdb, err := s.initBoardDB(board.Name)
		if err != nil {
			return err
		}
		boards[board.Name] = bdb
	}

	s.boardDBs = boards
	return nil
}

func (s *sqlite) Refresh() error {
	if err := s.Close(); err != nil {
		return err
	}
	return s.Init()
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

func (s *sqlite) PopulateBoard(boardName string, b *tchan.BoardOverview, ok *bool) error {
	boardDB, boardOK := s.boardDBs[boardName]
	if !boardOK {
		*ok = false
		return nil
	}

	bconf, confOK := s.conf.BoardConfig(boardName)
	if !confOK {
		return errors.Errorf("found DB but no config for /%s/", boardName)
	}
	*ok = true

	threadRows, err := boardDB.Query(`
SELECT t.topic, t.num_replies, t.created_at, t.active_at, op.id, op.author, op.content
FROM thread t INNER JOIN post op ON t.op_id = op.id
AND t.num_replies > -1 AND t.num_replies <= ?
ORDER BY t.active_at DESC
LIMIT ?;
`, bconf.MaxThreadLength, bconf.MaxThreadCount)
	if err != nil {
		return errors.Wrap(err, "failed to gather thread summaries")
	}
	defer threadRows.Close()

	b.Threads = make([]tchan.ThreadSummary, 0)
	for threadRows.Next() {
		t := tchan.ThreadSummary{}
		var createdTS, activeTS string
		err = threadRows.Scan(&t.Topic, &t.NumReplies, &createdTS, &activeTS,
			&t.OP.ID, &t.OP.Author, &t.OP.Content)
		if err != nil {
			errors.Wrap(err, "failed to extract thread summary")
		}

		var created time.Time
		if created, err = time.Parse(time.RFC3339, createdTS); err != nil {
			return errors.Wrap(err, "malformed date string in post table (created_at)")
		}
		t.OP.Timestamp = created

		var active time.Time
		if active, err = time.Parse(time.RFC3339, activeTS); err != nil {
			return errors.Wrap(err, "malformed date string in thread table (active_at)")
		}
		t.Active = active

		b.Threads = append(b.Threads, t)
	}

	return nil
}

func getThreadID(db *sql.DB, postID int64) (int64, bool, error) {
	var threadID int64
	result, err := db.Query(`
SELECT thread_id FROM post WHERE id = ?;
`, postID)
	if err != nil {
		return 0, false, err
	}
	defer result.Close()
	if !result.Next() {
		return threadID, false, nil
	}
	err = result.Scan(&threadID)
	if err != nil {
		return threadID, false, err
	}

	return threadID, true, nil
}

func getTopic(db *sql.DB, threadID int64) (string, error) {
	topic := ""
	result, err := db.Query(`
SELECT topic FROM thread WHERE id = ?;
`, threadID)
	if err != nil {
		return topic, err
	}
	defer result.Close()
	if !result.Next() {
		return topic, errors.Errorf("no thread table entry for thread %d", threadID)
	}

	err = result.Scan(&topic)
	if err != nil {
		return topic, errors.Errorf("invalid topic for thread %d", threadID)
	}

	return topic, nil
}

func (s *sqlite) PopulateThread(boardName string, postID int64, thr *tchan.Thread, ok *bool) error {
	*ok = false

	boardDB, boardOK := s.boardDBs[boardName]
	if !boardOK {
		return nil
	}

	threadID, idOK, err := getThreadID(boardDB, postID)
	if err != nil {
		return err
	}
	if !idOK {
		return nil
	}
	thr.Topic, err = getTopic(boardDB, threadID)
	if err != nil {
		return err
	}

	*ok = true

	result, err := boardDB.Query(`
SELECT id, author, created_at, content FROM post
WHERE thread_id = ?
ORDER BY created_at ASC;
`, threadID)
	if err != nil {
		return err
	}
	defer result.Close()

	for result.Next() {
		post := tchan.Post{}
		var ts string
		err = result.Scan(&post.ID, &post.Author, &ts, &post.Content)
		if err != nil {
			return err
		}

		post.Timestamp, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return err
		}
		thr.Posts = append(thr.Posts, post)
	}

	return nil
}

func (s *sqlite) CreateThread(boardName string, topic string, op *tchan.Post) error {
	boardDB, boardOK := s.boardDBs[boardName]
	if !boardOK {
		return errors.Errorf("attempting to create thread on non-existing board /%s/", boardName)
	}

	tresult, err := boardDB.Exec(`
INSERT INTO thread (topic) VALUES (?);
`, topic)
	if err != nil {
		return err
	}

	threadID, err := tresult.LastInsertId()
	if err != nil {
		return err
	}

	presult, err := boardDB.Exec(`
INSERT INTO post (thread_id, author, content) VALUES (?, ?, ?);
`, threadID, op.Author, op.Content)
	if err != nil {
		return err
	}

	op.ID, err = presult.LastInsertId()
	return err
}

func (s *sqlite) AddReply(boardName string, postID int64, post *tchan.Post, ok *bool) error {
	boardDB, boardOK := s.boardDBs[boardName]
	if !boardOK {
		return errors.Errorf("attempting to add post on non-existing board /%s/", boardName)
	}

	postRow, err := boardDB.Query(`
SELECT thread_id FROM post WHERE id = ?;
`, postID)
	if err != nil {
		return err
	}

	if !postRow.Next() {
		*ok = false
		return nil
	}
	*ok = true

	var threadID int
	if err = postRow.Scan(&threadID); err != nil {
		return err
	}

	if err = postRow.Close(); err != nil {
		return err
	}

	result, err := boardDB.Exec(`
INSERT INTO post (thread_id, author, content) VALUES (?, ?, ?);
`, threadID, post.Author, post.Content)
	if err != nil {
		return err
	}

	post.ID, err = result.LastInsertId()

	return err
}
