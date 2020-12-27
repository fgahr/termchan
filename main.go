package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fgahr/termchan/tchan/backend"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/http"
)

func run() error {
	var conf config.Opts

	var err error

	flag.StringVar(&conf.WorkingDirectory, "d", "./", "the base (configuration) directory for the service")
	flag.IntVar(&conf.Port, "p", 8088, "the port for the server to listen on")
	flag.Parse()

	if err = conf.Read(); err != nil {
		return err
	}

	db := backend.New(&conf)
	if err = db.Init(); err != nil {
		return err
	}

	srv := http.NewServer(&conf, db)
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

	return srv.ServeHTTP()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
