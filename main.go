package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fgahr/termchan/tchan2/backend"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/fgahr/termchan/tchan2/server"
)

func run() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

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
