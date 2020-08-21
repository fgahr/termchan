package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/backend"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/fgahr/termchan/tchan2/fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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
		f := fmt.GetWriter(r.URL.Query().Get("format"), w)
		boardName := mux.Vars(r)["board"]
		if !s.db.BoardExists(boardName) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		board, err := s.db.GetBoard(boardName)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = f.WriteBoard(board)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleViewThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f := fmt.GetWriter(r.URL.Query().Get("format"), w)
		boardName := mux.Vars(r)["board"]
		postID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			f.WriteError(errors.Errorf("invalid thread ID: %s", mux.Vars(r)["id"]))
			return
		}

		if !s.db.BoardExists(boardName) {
			w.WriteHeader(http.StatusNotFound)
			f.WriteError(errors.Errorf("no such board: %s", boardName))
			return
		}

		if !s.db.ThreadExists(boardName, postID) {
			w.WriteHeader(http.StatusNotFound)
			f.WriteError(errors.Errorf("no such post: %s/%d", boardName, postID))
			return
		}

		thread, err := s.db.GetThread(boardName, postID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			f.WriteError(errors.Errorf("internal server error"))
			return
		}

		err = f.WriteThread(thread)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func postParams(r *http.Request) (url.Values, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read request body")
	}
	values, err := url.ParseQuery(string(body))
	return values, errors.Wrap(err, "failed to parse request body")
}

func (s *Server) assemblePost(values url.Values) (tchan2.Post, error) {
	content := values.Get("content")
	// TODO:  Check content length

	author := "Anonymous"
	if name := values.Get("name"); name != "" {
		author = name
	}

	return tchan2.Post{Author: author, ID: -1, Timestamp: time.Now(), Content: content}, nil
}

func (s *Server) handleCreateThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f := fmt.GetWriter("", w)
		boardName := mux.Vars(r)["board"]
		if !s.db.BoardExists(boardName) {
			w.WriteHeader(http.StatusNotFound)
			f.WriteError(errors.Errorf("no such board: %s", boardName))
			return
		}

		params, err := postParams(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		f = fmt.GetWriter(params.Get("format"), w)

		post, err := s.assemblePost(params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			f.WriteError(errors.New("unable to parse post from request data"))
			return
		}

		topic := params.Get("topic")

		id, err := s.db.CreateThread(boardName, topic, post)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			f.WriteError(errors.New("failed to create thread"))
			return
		}

		thread, err := s.db.GetThread(boardName, id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			f.WriteError(errors.New("failed to fetch thread after creating it"))
			return
		}

		err = f.WriteThread(thread)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleReplyToThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := newRequestWorker(w, r, s.conf)
		rw.extractPost()
		// TODO

		// postID, err := strconv.Atoi(mux.Vars(r)["id"])
		// if err != nil {
		// 	log.Printf("failed to get post ID from mux.Vars: %v", mux.Vars(r))
		// 	w.WriteHeader(http.StatusBadRequest)

		// }

		// id, err := s.db.AddReply(boardName, postID, post)
		// if err != nil {
		// 	log.Println(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// }

		// thread, err := s.db.GetThread(boardName, id)
		// if err != nil {
		// 	log.Println(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	f.WriteError(errors.New("failed to fetch thread after creating it"))
		// 	return
		// }

		// err = f.WriteThread(thread)
		// if err != nil {
		// 	log.Println(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// }
	}
}
