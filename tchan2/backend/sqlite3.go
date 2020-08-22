package backend

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/fgahr/termchan/tchan2/util"
)

type sqlite struct {
	baseDirectory   string
	boardsDirectory string
	boardDBs        map[string]*sql.DB
}

func (s *sqlite) Init() error {
	var err error
	var ok bool

	if ok, err = util.DirExists(s.baseDirectory); err != nil {
		return err
	} else if !ok {
		err = os.Mkdir(s.baseDirectory, 0700)
	}

	return nil
}

func (s *sqlite) initBoardDB(boardName string) error {
	path := filepath.Join(s.boardsDirectory, boardName+".db")
	boardDB, err := sql.Open("sqlite3", path)
	s.boardDBs[boardName] = boardDB
	return err
}
