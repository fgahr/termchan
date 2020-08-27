package server

import (
	"net/http"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/backend"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/gorilla/mux"
)

// Server connects all aspects of the termchan application.
type Server struct {
	conf   *config.Opts
	db     backend.DB
	router *mux.Router
}

// New creates a new server without configuration or backend.
// In order to be usable these still need to be set up.
func New(opts *config.Opts, db backend.DB) *Server {
	s := &Server{conf: opts, db: db, router: mux.NewRouter()}
	s.routes()
	return s
}

// ServeHTTP handles HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleWelcome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)
		rw.respondWelcome()
	}
}

func (s *Server) handleViewBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func (s *Server) handleViewThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func (s *Server) handleCreateThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
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
