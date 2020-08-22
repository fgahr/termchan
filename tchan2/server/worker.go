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
			log.Println(err)
			rw.err = errors.New("unable to read request body")
			// TODO: Check which HTTP status is appropriate
			rw.respondError(http.StatusPreconditionFailed)
			return
		}
		rw.params, rw.err = url.ParseQuery(string(body))
	default:
		rw.err = errors.Errorf("illegal request method: %s", rw.r.Method)
		log.Println(rw.err)
		rw.respondError(http.StatusBadRequest)
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
	if _, ok := rw.conf.BoardConfig(rw.board); !ok {
		rw.err = errors.Errorf("no such board: %s", rw.board)
		rw.respondError(http.StatusNotFound)
		return
	}

	id := vars["id"]
	if id == "" {
		rw.replyID = 0
	} else {
		rw.replyID, rw.err = strconv.Atoi(id)
	}

	if rw.err != nil {
		rw.err = errors.Errorf("invalid post ID: %s", id)
		rw.respondError(http.StatusInternalServerError)
		return
	}
}

func (rw *requestWorker) try(f func() error, failStatus int, errorText string) {
	if rw.err != nil {
		return
	}

	err := f()
	log.Println(err)
	if err != nil {
		if errorText != "" {
			err = errors.New(errorText)
		}
		rw.err = err
		rw.respondError(failStatus)
	}
}

func (rw *requestWorker) extractPost() {
	// Trimming extraneous spaces avoids all kinds of abuse
	content := strings.TrimSpace(rw.params.Get("content"))
	bc, ok := rw.conf.BoardConfig(rw.board)
	if !ok {
		rw.err = errors.Errorf("no such board: %s", rw.board)
		rw.respondError(http.StatusNotFound)
		return
	}
	if len(content) > bc.MaxPostBytes {
		rw.err = errors.Errorf("post too large: %d bytes (max %d bytes)", len(content), bc.MaxPostBytes)
		rw.respondError(http.StatusBadRequest)
		return
	} else if content == "" {
		rw.err = errors.Errorf("empty post content")
		rw.respondError(http.StatusBadRequest)
		return
	}

	author := "Anonymous"
	if name := rw.params.Get("name"); name != "" {
		author = name
	}

	rw.post = tchan2.Post{Author: author, Timestamp: time.Now(), Content: content}
}

func (rw *requestWorker) getTopic() string {
	if rw.err != nil {
		return ""
	}
	return rw.params.Get("topic")
}

func (rw *requestWorker) respondThread(thr tchan2.Thread) {
	if rw.err != nil {
		return
	}

	rw.err = rw.f.WriteThread(thr)
	if rw.err != nil {
		// No point in trying to print anything else to the client
		// just set the status code
		rw.w.WriteHeader(http.StatusInternalServerError)
		log.Println(rw.err)
	}
}

func (rw *requestWorker) respondBoard(b tchan2.BoardOverview) {
	if rw.err != nil {
		return
	}

	rw.err = rw.f.WriteBoard(b)
	if rw.err != nil {
		// No point in trying to print anything else to the client
		// just set the status code
		rw.w.WriteHeader(http.StatusInternalServerError)
		log.Println(rw.err)
	}
}

func (rw *requestWorker) respondError(statusCode int) {
	rw.w.WriteHeader(statusCode)
	rw.f.WriteError(rw.err)
}
