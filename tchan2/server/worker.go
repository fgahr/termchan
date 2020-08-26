package server

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fgahr/termchan/tchan2"
	"github.com/fgahr/termchan/tchan2/config"
	"github.com/fgahr/termchan/tchan2/fmt"
	"github.com/fgahr/termchan/tchan2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type requestWorker struct {
	w       http.ResponseWriter
	r       *http.Request
	conf    *config.Opts
	f       fmt.Writer
	params  url.Values
	board   string
	replyID int64
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
			log.Error(err)
			rw.err = errors.New("unable to read request body")
			// TODO: Check which HTTP status is appropriate
			rw.respondError(http.StatusPreconditionFailed)
			return
		}
		rw.params, rw.err = url.ParseQuery(string(body))
	default:
		rw.err = errors.Errorf("illegal request method: %s", rw.r.Method)
		log.Error(rw.err)
		rw.respondError(http.StatusBadRequest)
	}
}

func (rw *requestWorker) setUpWriter() {
	if rw.err != nil {
		return
	}

	rw.f = fmt.GetWriter(rw.params, rw.r, rw.w)
}

func (rw *requestWorker) determineBoardAndPost() {
	if rw.err != nil {
		return
	}

	vars := mux.Vars(rw.r)
	rw.board = vars["board"]

	id := vars["id"]
	if id == "" {
		rw.replyID = 0
	} else {
		rw.replyID, rw.err = strconv.ParseInt(id, 10, 64)
	}

	if rw.err != nil {
		rw.err = errors.Errorf("invalid post ID: %s", id)
		rw.respondError(http.StatusInternalServerError)
		return
	}
}

func (rw *requestWorker) try(f func() error, failStatus int, errorText string, handlers ...func(error)) {
	if rw.err != nil {
		return
	}

	err := f()
	if err != nil {
		log.Error(err)
		if errorText != "" {
			err = errors.New(errorText)
		}
		rw.err = err
		rw.respondError(failStatus)
	}
}

func (rw *requestWorker) extractPost() {
	if rw.err != nil {
		return
	}

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

func (rw *requestWorker) respondWelcome() {
	rw.try(func() error { return rw.f.WriteWelcome(rw.conf.Boards) },
		http.StatusInternalServerError, "", log.Error)
}

func (rw *requestWorker) respondThread(thr tchan2.Thread) {
	rw.try(func() error { return rw.f.WriteThread(thr) },
		http.StatusInternalServerError, "", log.Error)
}

func (rw *requestWorker) respondNoSuchThread() {
	if rw.err != nil {
		return
	}

	rw.err = errors.Errorf("no such thread: /%s/%d", rw.board, rw.replyID)
	rw.respondError(http.StatusNotFound)
}

func (rw *requestWorker) respondBoard(b tchan2.BoardOverview) {
	rw.try(func() error { return rw.f.WriteBoard(b) },
		http.StatusInternalServerError, "", log.Error)
}

func (rw *requestWorker) respondNoSuchBoard() {
	if rw.err != nil {
		return
	}

	rw.err = errors.Errorf("no such board: /%s/", rw.board)
	rw.respondError(http.StatusNotFound)
}

func (rw *requestWorker) respondError(statusCode int) {
	rw.w.WriteHeader(statusCode)
	rw.f.WriteError(rw.err)
}
