package http

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/backend"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/output"
	"github.com/fgahr/termchan/tchan/output/ansi"
	"github.com/fgahr/termchan/tchan/output/html"
	"github.com/fgahr/termchan/tchan/output/json"
	"github.com/fgahr/termchan/tchan/util"
)

// Server connects all aspects of the termchan application.
type Server struct {
	conf     *config.Settings
	hs       *http.Server
	db       backend.DB
	router   *mux.Router
	confLock *sync.RWMutex
	htmlSet  html.TemplateSet
	ansiSet  ansi.TemplateSet
}

// New creates a new server with configuration and backend.
// Backend is assumed to be fully set up.
func NewServer(conf *config.Settings) (*Server, error) {
	db := backend.New(conf)
	if err := db.Init(); err != nil {
		return nil, errors.Wrap(err, "backend setup failed")
	}

	s := &Server{
		conf:     conf,
		db:       db,
		router:   mux.NewRouter(),
		confLock: new(sync.RWMutex),
	}
	s.routes()

	if err := s.ReloadConfig(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.handleWelcome()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}", s.handleViewBoard()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/", s.handleViewBoard()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}", s.handleCreateThread()).Methods("POST")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/", s.handleCreateThread()).Methods("POST")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/{id:[0-9]+}", s.handleViewThread()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/{id:[0-9]+}/", s.handleViewThread()).Methods("GET")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/{id:[0-9]+}", s.handleReplyToThread()).Methods("POST")
	s.router.HandleFunc("/{board:[a-zA-Z0-9]+}/{id:[0-9]+}/", s.handleReplyToThread()).Methods("POST")
}

// ReloadConfig forces the server to reload its configuration and templates.
// New connections are stalled until the process is completed.
func (s *Server) ReloadConfig() error {
	s.confLock.Lock()
	defer s.confLock.Unlock()

	log.Println("loading configuration")
	if err := s.conf.ReadFromFile(); err != nil {
		return err
	}

	log.Println("reading templates")
	if err := s.htmlSet.Read(s.conf.TemplateDirectory()); err != nil {
		return errors.Wrap(err, "reading html templates failed")
	}

	if err := s.ansiSet.Read(s.conf.TemplateDirectory()); err != nil {
		return errors.Wrap(err, "reading ansi templates failed")
	}

	return s.db.Refresh()
}

// ServeHTTP handles HTTP requests.
func (s *Server) ServeHTTP() error {
	t := s.conf.Transport
	if t.Protocol == config.Unix {
		if exists, err := util.FileExists(t.Socket); err != nil {
			return errors.Wrapf(err, "unable to check status of socket file%s", t.Socket)
		} else if exists {
			return errors.Errorf("cannot open socket: file %s exists", t.Socket)
		} else {
			// Clean it up after we're done
			defer os.Remove(t.Socket)
		}
	}

	listener, err := net.Listen(t.Protocol.String(), t.Socket)
	if err != nil {
		return errors.Wrapf(err, "unable to establish listener on %v", t)
	}

	if t.Protocol == config.Unix {
		if err := os.Chmod(t.Socket, 0666); err != nil {
			return errors.Wrapf(err, "unable to open socket %s for other services", t.Socket)
		}
	}

	s.hs = &http.Server{
		Addr:    t.Socket,
		Handler: s.router,
	}
	log.Printf("serving HTTP on %v", s.conf.Transport)
	if err := s.hs.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop causes the server to stop listening.
func (s *Server) Stop() error {
	if s.hs != nil {
		return s.hs.Shutdown(context.Background())
	}
	return errors.New("not listening")
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
		rw := s.newRequestWorker(w, r)
		rw.respondWelcome()
	})
}

func (s *Server) handleViewBoard() http.HandlerFunc {
	return s.confReader(func(w http.ResponseWriter, r *http.Request) {
		rw := s.newRequestWorker(w, r)

		boardConf, ok := s.conf.BoardConfig(rw.board)
		if !ok {
			rw.respondNoSuchBoard()
		}

		ok = false
		board := tchan.BoardOverview{Board: boardConf}
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
		rw := s.newRequestWorker(w, r)

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
		rw := s.newRequestWorker(w, r)

		rw.extractPost()
		topic := rw.getTopic()

		rw.try(func() error { return s.db.CreateThread(rw.board, topic, &rw.post) },
			http.StatusInternalServerError, "failed to create thread")

		boardConf, ok := s.conf.BoardConfig(rw.board)
		if !ok {
			rw.respondNoSuchBoard()
		}
		thr := tchan.Thread{Board: boardConf}

		ok = false
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
		rw := s.newRequestWorker(w, r)

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

func (s *Server) jsonWriter(r *http.Request, w http.ResponseWriter) output.Writer {
	return json.NewWriter(r, w)
}

func (s *Server) ansiWriter(r *http.Request, w http.ResponseWriter) output.Writer {
	return ansi.NewWriter(r, w, s.ansiSet)
}

func (s *Server) htmlWriter(r *http.Request, w http.ResponseWriter) output.Writer {
	return html.NewWriter(r, w, s.htmlSet)
}
