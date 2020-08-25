package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/fgahr/termchan/tchan2/backend"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/fgahr/termchan/tchan2/server"
)

func run() error {
	var err error
	var wd string
	flag.StringVar(&wd, "dir", "./", "the base (configuration) directory for the service")
	flag.Parse()

	conf := config.New(wd)
	if err = conf.Read(); err != nil {
		return err
	}

	db := backend.New(conf)
	if err = db.Init(); err != nil {
		return err
	}

	srv := server.New(conf, db)

	return http.ListenAndServe(":8088", srv)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
