package tchan2

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Config deals with all variable and optional aspects of termchan.
type Config struct {
	// Working directory
	PWD string
}

// Backend handles all database interactions.
type Backend interface {
	// TODO
}

// Server connects all aspects of the termchan application.
type Server struct {
	Conf   *Config
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
	// TODO
}
