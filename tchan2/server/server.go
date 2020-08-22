package server

import (
	"log"
	"net/http"

	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/backend"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/fgahr/termchan/tchan2/fmt"
	"github.com/gorilla/mux"
)

// SelectWriter chooses an appropriate writer for the given request.
func SelectWriter(w http.ResponseWriter, r *http.Request) fmt.Writer {
	format := r.URL.Query().Get("format")
	return fmt.GetWriter(format, w)
}

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
		f := fmt.GetWriter(r.URL.Query().Get("format"), w)
		err := f.WriteWelcome()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleListBoards() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writer := fmt.GetWriter(r.URL.Query().Get("format"), w)
		err := writer.WriteOverview(s.conf.Boards)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleViewBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		b := tchan2.BoardOverview{}
		rw.try(func() error {
			return s.db.PopulateBoard(rw.board, &b)
		}, http.StatusInternalServerError, "failed to fetch board")

		rw.respondBoard(b)
	}
}

func (s *Server) handleViewThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)

		thr := tchan2.Thread{}
		rw.try(func() error { return s.db.PopulateThread(rw.board, rw.replyID, &thr) },
			http.StatusInternalServerError, "failed to fetch thread for viewing")

		rw.respondThread(thr)
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
		rw.try(func() error { return s.db.PopulateThread(rw.board, rw.post.ID, &thr) },
			http.StatusInternalServerError, "failed to fetch thread for viewing")

		rw.respondThread(thr)
	}
}

func (s *Server) handleReplyToThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)
		rw.extractPost()

		rw.try(func() error { return s.db.AddAsReply(rw.board, rw.replyID, &rw.post) },
			http.StatusInternalServerError, "failed to persist reply")

		thr := tchan2.Thread{}
		rw.try(func() error { return s.db.PopulateThread(rw.board, rw.replyID, &thr) },
			http.StatusInternalServerError, "failed to fetch thread for viewing")

		rw.respondThread(thr)
	}
}
