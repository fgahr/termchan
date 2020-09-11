package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/backend"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/gorilla/mux"
)

// Server connects all aspects of the termchan application.
type Server struct {
	conf     *config.Opts
	db       backend.DB
	router   *mux.Router
	confLock *sync.RWMutex
}

// New creates a new server without configuration or backend.
// In order to be usable these still need to be set up.
func New(opts *config.Opts, db backend.DB) *Server {
	s := &Server{
		conf:     opts,
		db:       db,
		router:   mux.NewRouter(),
		confLock: new(sync.RWMutex),
	}
	s.routes()
	return s
}

func (s *Server) ReloadConfig() error {
	s.confLock.Lock()
	defer s.confLock.Unlock()

	log.Println("reloading configuration")
	if err := s.conf.Read(); err != nil {
		return err
	}

	return s.db.Refresh()
}

// ServeHTTP handles HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) confReader(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.confLock.RLock()
		defer s.confLock.RUnlock()
		f(w, r)
	}
}

func (s *Server) handleWelcome() http.HandlerFunc {
	return s.confReader(func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)
		rw.respondWelcome()
	})
}

func (s *Server) handleViewBoard() http.HandlerFunc {
	return s.confReader(func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		boardConf, ok := s.conf.BoardConfig(rw.board)
		if !ok {
			rw.respondNoSuchBoard()
		}

		ok = false
		board := tchan.BoardOverview{MetaData: boardConf}
		rw.try(func() error {
			return s.db.PopulateBoard(rw.board, &board, &ok)
		}, http.StatusInternalServerError, "failed to fetch board")

		if ok {
			rw.respondBoard(board)
		} else {
			rw.respondNoSuchBoard()
		}
	})
}

func (s *Server) handleViewThread() http.HandlerFunc {
	return s.confReader(func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		boardConf, ok := s.conf.BoardConfig(rw.board)
		if !ok {
			rw.respondNoSuchBoard()
		}
		thr := tchan.Thread{Board: boardConf}

		ok = false
		rw.try(func() error { return s.db.PopulateThread(rw.board, rw.replyID, &thr, &ok) },
			http.StatusInternalServerError, "failed to fetch thread for viewing")

		if ok {
			rw.respondThread(thr)
		} else {
			rw.respondNoSuchThread()
		}
	})
}

func (s *Server) handleCreateThread() http.HandlerFunc {
	return s.confReader(func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		rw.extractPost()
		topic := rw.getTopic()

		rw.try(func() error { return s.db.CreateThread(rw.board, topic, &rw.post) },
			http.StatusInternalServerError, "failed to create thread")

		thr := tchan.Thread{}
		ok := false
		rw.try(func() error { return s.db.PopulateThread(rw.board, rw.post.ID, &thr, &ok) },
			http.StatusInternalServerError, "failed to fetch thread for viewing")

		if ok {
			rw.respondThread(thr)
		} else {
			rw.respondNoSuchThread()
		}
	})
}

func (s *Server) handleReplyToThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		boardConf, ok := s.conf.BoardConfig(rw.board)
		if !ok {
			rw.respondNoSuchBoard()
		}

		rw.extractPost()
		ok = false
		rw.try(func() error { return s.db.AddReply(rw.board, rw.replyID, &rw.post, &ok) },
			http.StatusInternalServerError, "failed to persist reply")
		if !ok {
			rw.respondNoSuchThread()
		}

		thr := tchan.Thread{Board: boardConf}
		rw.try(func() error { return s.db.PopulateThread(rw.board, rw.replyID, &thr, &ok) },
			http.StatusInternalServerError, "failed to fetch thread for viewing")

		if ok {
			rw.respondThread(thr)
		} else {
			rw.respondNoSuchThread()
		}
	}
}
