package server

import (
	"net/http"

	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/backend"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/gorilla/mux"
)

// Server connects all aspects of the termchan application.
type Server struct {
	conf   *config.Opts
	db     backend.DB
	router *mux.Router
}

// NewServer creates a new server without configuration or backend.
// In order to be usable these still need to be set up.
func NewServer(opts *config.Opts, db backend.DB) *Server {
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

func (s *Server) handleListBoards() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)
		rw.respondBoardList()
	}
}

func (s *Server) handleViewBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		b := tchan2.BoardOverview{}
		ok := false
		rw.try(func() error {
			return s.db.PopulateBoard(rw.board, &b, &ok)
		}, http.StatusInternalServerError, "failed to fetch board")

		if ok {
			rw.respondBoard(b)
		} else {
			rw.respondNoSuchBoard()
		}
	}
}

func (s *Server) handleViewThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		thr := tchan2.Thread{}
		ok := false
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

		thr := tchan2.Thread{}
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
		rw.extractPost()

		ok := false
		rw.try(func() error { return s.db.AddAsReply(rw.board, rw.replyID, &rw.post, &ok) },
			http.StatusInternalServerError, "failed to persist reply")
		if !ok {
			rw.respondNoSuchThread()
		}

		thr := tchan2.Thread{}
		rw.try(func() error { return s.db.PopulateThread(rw.board, rw.replyID, &thr, &ok) },
			http.StatusInternalServerError, "failed to fetch thread for viewing")

		if ok {
			rw.respondThread(thr)
		} else {
			rw.respondNoSuchThread()
		}
	}
}
