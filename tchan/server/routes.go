package server

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
