package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fgahr/termchan/tchan/backend"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/server"
)

func run() error {
	var err error
	var wd string
	var port int

	flag.StringVar(&wd, "d", "./", "the base (configuration) directory for the service")
	flag.IntVar(&port, "p", 8088, "the port for the server to listen on")
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
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	defer close(sigChan)
	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGHUP:
				srv.ReloadConfig()
			default:
				panic("unexpected signal")
			}
		}
	}()

	log.Printf("serving HTTP on port %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), srv)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
