package http

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/fgahr/termchan/tchan"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/output"
)

type requestWorker struct {
	conf    *config.Settings
	w       output.Writer
	r       *http.Request
	params  url.Values
	board   string
	replyID int64
	post    tchan.Post
	err     error
}

func (s *Server) newRequestWorker(w http.ResponseWriter, r *http.Request) *requestWorker {
	// Use ANSI as default for possible error messages up to this point.
	rw := requestWorker{conf: s.conf, w: s.ansiWriter(r, w), r: r}
	rw.init()
	switch rw.params.Get("format") {
	case "ansi":
		rw.w = s.ansiWriter(r, w)
	case "html":
		rw.w = s.htmlWriter(r, w)
	case "json":
		rw.w = s.jsonWriter(r, w)
	default:
		rw.w = s.ansiWriter(r, w)
	}

	return &rw
}

func (rw *requestWorker) init() {
	rw.readParams()
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
			rw.respondError(http.StatusInternalServerError)
			return
		}
		rw.params, rw.err = url.ParseQuery(string(body))
	default:
		rw.err = errors.Errorf("illegal request method: %s", rw.r.Method)
		log.Println(rw.err)
		rw.respondError(http.StatusBadRequest)
	}
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
		log.Println(err)
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

	// Trimming extraneous spaces avoids some kinds of abuse/trolling
	content := strings.TrimSpace(rw.params.Get("content"))
	bc, ok := rw.conf.BoardConfig(rw.board)
	if !ok {
		rw.err = errors.Errorf("no such board: %s", rw.board)
		rw.respondError(http.StatusNotFound)
		return
	}
	if len(content) > bc.MaxPostBytes() {
		rw.err = errors.Errorf("post too large: %d bytes (max %d bytes)", len(content), bc.MaxPostBytes())
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

	rw.post = tchan.Post{Author: author, Timestamp: time.Now(), Content: content}
}

func (rw *requestWorker) getTopic() string {
	if rw.err != nil {
		return ""
	}
	return rw.params.Get("topic")
}

func (rw *requestWorker) respondWelcome() {
	rw.try(func() error { return rw.w.WriteWelcome(rw.conf.Boards) },
		http.StatusInternalServerError, "", func(err error) { log.Println(err) })
}

func (rw *requestWorker) respondThread(thr tchan.Thread) {
	rw.try(func() error { return rw.w.WriteThread(thr) },
		http.StatusInternalServerError, "", func(err error) { log.Println(err) })
}

func (rw *requestWorker) respondNoSuchThread() {
	if rw.err != nil {
		return
	}

	rw.err = errors.Errorf("no such thread: /%s/%d", rw.board, rw.replyID)
	rw.respondError(http.StatusNotFound)
}

func (rw *requestWorker) respondBoard(b tchan.BoardOverview) {
	rw.try(func() error { return rw.w.WriteBoard(b) },
		http.StatusInternalServerError, "", func(err error) { log.Println(err) })
}

func (rw *requestWorker) respondNoSuchBoard() {
	if rw.err != nil {
		return
	}

	rw.err = errors.Errorf("no such board: /%s/", rw.board)
	rw.respondError(http.StatusNotFound)
}

func (rw *requestWorker) respondError(status int) {
	rw.w.WriteError(status, rw.err)
}
