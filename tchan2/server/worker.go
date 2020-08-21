package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/fgahr/termchan/tchan2/fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type requestWorker struct {
	w http.ResponseWriter
	r *http.Request
	// db      backend.DB
	conf    *config.Opts
	f       fmt.Writer
	params  url.Values
	board   string
	replyID int
	post    tchan2.Post
	err     error
}

func newRequestWorker(w http.ResponseWriter, r *http.Request, opts *config.Opts) *requestWorker {
	rw := requestWorker{w: w, r: r, conf: opts}
	rw.init()
	return &rw
}

func (rw *requestWorker) init() {
	rw.readParams()
	rw.setUpWriter()
	rw.determineBoardAndPost()
}

func (rw *requestWorker) readParams() {
	if rw.err != nil {
		return
	}

	switch rw.r.Method {
	case "GET":
		rw.params = rw.r.URL.Query()
	case "POST":
		body, err := ioutil.ReadAll(rw.r.Body)
		if err != nil {
			err = errors.Wrap(err, "unable to read request body")
		}
		rw.params, rw.err = url.ParseQuery(string(body))
	default:
		rw.err = errors.Errorf("illegal request method: %s", rw.r.Method)
		rw.w.WriteHeader(http.StatusBadRequest)
		log.Println(rw.err)
	}
}

func (rw *requestWorker) setUpWriter() {
	if rw.err != nil {
		return
	}

	rw.f = fmt.GetWriter(rw.params.Get("format"), rw.w)
}

func (rw *requestWorker) determineBoardAndPost() {
	if rw.err != nil {
		return
	}

	vars := mux.Vars(rw.r)
	rw.board = vars["board"]
	if !rw.conf.BoardExists(rw.board) {
		rw.w.WriteHeader(http.StatusNotFound)
		rw.err = errors.Errorf("no such board: %s", rw.board)
		rw.f.WriteError(rw.err)
		return
	}

	id := vars["id"]
	if id == "" {
		rw.replyID = 0
	} else {
		rw.replyID, rw.err = strconv.Atoi(id)
	}

	if rw.err != nil {
		rw.w.WriteHeader(http.StatusInternalServerError)
		rw.err = errors.Errorf("invalid post ID: %s", id)
		rw.f.WriteError(rw.err)
		return
	}
}

func (rw *requestWorker) extractPost() {
	// Trimming extraneous spaces avoids all kinds of abuse
	content := strings.TrimSpace(rw.params.Get("content"))
	bc := rw.conf.BoardConfig(rw.board)
	if len(content) > bc.MaxPostBytes {
		rw.w.WriteHeader(http.StatusBadRequest)
		rw.err = errors.Errorf("post too large: %d bytes (max %d bytes)", len(content), bc.MaxPostBytes)
		rw.f.WriteError(rw.err)
		return
	} else if content == "" {
		rw.w.WriteHeader(http.StatusBadRequest)
		rw.err = errors.Errorf("empty post content")
		rw.f.WriteError(rw.err)
	}

	author := "Anonymous"
	if name := rw.params.Get("name"); name != "" {
		author = name
	}

	rw.post = tchan2.Post{Author: author, Timestamp: time.Now(), Content: content}
}
