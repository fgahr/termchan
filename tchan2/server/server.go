package server

import (
	"net/http"

	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/gorilla/mux"
)

// TODO: Move to a more appropriate place
//
// Writer describes an entity in charge of writing a server response.
type Writer interface {
	WriteWelcome(boardData []tchan2.BoardMetaData) error
	WriteThread(thread tchan2.ThreadFull) error
	WriteBoard(board tchan2.BoardOverview) error
	WriteError(err error) error
}

// SelectWriter chooses an appropriate writer for the given request.
func SelectWriter(w http.ResponseWriter, r *http.Request) Writer {
	// TODO
	return nil
}

// Backend handles all database interactions.
type Backend interface {
	GetBoardMetaData(boardName string) (tchan2.BoardMetaData, error)
	GetBoardOverview(boardName string) error
	GetThread(boardName string, threadID int) error
}

// Server connects all aspects of the termchan application.
type Server struct {
	Conf   *config.Opts
	DB     Backend
	router *mux.Router
}

// NewServer creates a new server without configuration or backend.
// In order to be usable these still need to be set up.
func NewServer() *Server {
	s := &Server{router: mux.NewRouter()}
	s.routes()
	return s
}

// ServerHTTP handles HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.handleWelcome()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}", s.handleViewBoard()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/", s.handleViewBoard()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}", s.handleCreateThread()).Methods("POST")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/", s.handleCreateThread()).Methods("POST")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+/{id:[0-9]+}", s.handleViewThread()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+/{id:[0-9]+}/", s.handleViewThread()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+/{id:[0-9]+}", s.handleReplyToThread()).Methods("POST")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+/{id:[0-9]+}/", s.handleReplyToThread()).Methods("POST")
}

func (s *Server) handleWelcome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}

func (s *Server) handleViewBoard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}

func (s *Server) handleViewThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}

func (s *Server) handleCreateThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}

func (s *Server) handleReplyToThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}
}
