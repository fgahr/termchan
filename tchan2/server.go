package tchan2

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Config deals with all variable and optional aspects of termchan.
type Config struct {
	WorkingDirectory string
}

// Backend handles all database interactions.
type Backend interface {
	GetBoardMetaData(boardName string) (BoardMetaData, error)
	GetBoardOverview(boardName string) error
	GetThread(boardName string, threadID int) error
}

// Server connects all aspects of the termchan application.
type Server struct {
	Conf   *Config
	DB     Backend
	router *mux.Router
}

// Formatter describes an entity in charge of formatting a server response.
type Formatter interface {
	FormatWelcome()
	FormatThread()
	FormatBoard()
	FormatError()
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
