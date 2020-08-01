package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fgahr/termchan/tchan"
	. "github.com/fgahr/termchan/tchan/config"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func main() {
	var err error

	if Conf.LogFile != "" {
		if f, err := os.OpenFile(Conf.LogFile, os.O_APPEND, 0600); err != nil {
			log.Fatalf("unable to open log file %s", Conf.LogFile)
		} else {
			defer f.Close()
			log.SetOutput(f)
		}
	}

	if err = tchan.Initialize(); err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize termchan"))
	}

	router := mux.NewRouter()
	router.HandleFunc("/", tchan.HandleWelcome).Methods(http.MethodGet)
	router.HandleFunc("/{board:[a-z]+}", tchan.ViewBoard).Methods(http.MethodGet)
	router.HandleFunc("/{board:[a-z]+}/", tchan.ViewBoard).Methods(http.MethodGet)
	router.HandleFunc("/{board:[a-z]+}", tchan.CreateThread).Methods(http.MethodPost)
	router.HandleFunc("/{board:[a-z]+}/", tchan.CreateThread).Methods(http.MethodPost)
	router.HandleFunc("/{board:[a-z]+}/{id:[0-9]+}", tchan.ViewThread).Methods(http.MethodGet)
	router.HandleFunc("/{board:[a-z]+}/{id:[0-9]+}", tchan.ReplyToThread).Methods(http.MethodPost)

	log.Printf("Serving HTTP on %s\n", Conf.PortString())
	err = http.ListenAndServe(Conf.PortString(), router)
	if err != nil {
		log.Println(err)
	}

	err = tchan.Shutdown()
	if err != nil {
		log.Fatal(err)
	}
}
