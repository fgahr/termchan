package tchan

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/fgahr/termchan/tchan/backend"
	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/data"
	"github.com/fgahr/termchan/tchan/format"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Initialize starts the backend and collects runtime information.
func Initialize() error {
	var err error

	if err = backend.Connect(); err != nil {
		return errors.Wrap(err, "failed to connect to backend")
	}

	data.Boards, err = backend.FetchBoardIDs()
	data.BoardParams = data.GatherBoardParameters()
	return err
}

// Shutdown handles backend shutdown and associated actions.
func Shutdown() error {
	return backend.Shutdown()
}

// HandleWelcome returns a welcome message.
func HandleWelcome(w http.ResponseWriter, r *http.Request) {
	f := format.Select(r.URL.Query(), w)
	f.FormatWelcome(data.BoardParams, r.Host)
}

// ViewBoard returns a board overview.
func ViewBoard(w http.ResponseWriter, r *http.Request) {
	f := format.Select(r.URL.Query(), w)
	bname := mux.Vars(r)["board"]
	if board, ok := data.Boards[bname]; ok {
		b, err := backend.GetBoard(board.ID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			f.FormatError(errors.Errorf("failed to fetch board: /%s/", bname))
		}
		f.FormatBoard(b)
	} else {
		w.WriteHeader(http.StatusNotFound)
		f.FormatError(errors.Errorf("no such board: /%s/", bname))
	}
}

// Finds the post from the in-board ID used in the request.
func boardPostIDFromRequest(r *http.Request) (string, data.LPid, bool) {
	board := mux.Vars(r)["board"]
	idField := mux.Vars(r)["id"]
	rawPid, err := strconv.Atoi(idField)
	pid := data.LPid(rawPid)
	if err != nil {
		return board, pid, false
	}
	return board, pid, true
}

// ViewThread returns a thread overview.
func ViewThread(w http.ResponseWriter, r *http.Request) {
	f := format.Select(r.URL.Query(), w)
	bname, opID, ok := boardPostIDFromRequest(r)
	board, ok := data.Boards[bname]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		f.FormatError(errors.Errorf("no such board: /%s/", bname))
		return
	}

	tID, ok, err := backend.GetThreadID(board.ID, opID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.Errorf("couldn't find thread: /%s/%d", bname, opID))
		return
	} else if !ok {
		w.WriteHeader(http.StatusNotFound)
		f.FormatError(errors.Errorf("no such thread: /%s/%d", bname, opID))
		return
	}

	thread, err := backend.GetThread(tID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.Errorf("failed to fetch thread: /%s/%d", bname, opID))
		return
	}
	thread.Board = board

	f.FormatThread(thread)
}

func parsePostParameters(r *http.Request) (url.Values, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read request body")
	}
	values, err := url.ParseQuery(string(body))
	return values, errors.Wrap(err, "failed to parse request body")
}

//  Gathers relevant data for creating a new post.
func assembleNewPost(values url.Values) (data.Post, error) {
	content := values.Get("content")
	if len(content) > config.Current.Max.PostSize {
		return data.Post{}, errors.Errorf("post content exceeds %d byte limit", config.Current.Max.PostSize)
	}

	author := "Anonymous"
	if name := values.Get("name"); name != "" {
		author = name
	}
	return data.Post{Author: author, AuthorIP: "", Content: content, Timestamp: time.Now()}, nil
}

// CreateThread handles the creation of new threads on request.
func CreateThread(w http.ResponseWriter, r *http.Request) {
	params, err := parsePostParameters(r)
	f := format.Select(params, w)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		// TODO: wrap error
		f.FormatError(err)
	}

	bname, _, _ := boardPostIDFromRequest(r)
	board, ok := data.Boards[bname]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		f.FormatError(errors.Errorf("no such board: /%s/", bname))
		return
	}

	topic := params.Get("topic")
	tID, err := backend.CreateThread(board.ID, topic)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.New("unable to create thread"))
		return
	}

	post, err := assembleNewPost(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		f.FormatError(err)
		return
	}

	err = backend.AddReplyToThread(tID, &post)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.New("unable to create thread"))
		return
	}

	thread, err := backend.GetThread(tID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.New("unable to create thread"))
		return
	}
	thread.Board = board

	f.FormatThread(thread)
}

// ReplyToThread handles responses to existing threads.
func ReplyToThread(w http.ResponseWriter, r *http.Request) {
	params, err := parsePostParameters(r)
	f := format.Select(params, w)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		// TODO: wrap error
		f.FormatError(err)
	}

	bname, opID, ok := boardPostIDFromRequest(r)
	board, ok := data.Boards[bname]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		f.FormatError(errors.Errorf("no such board: /%s/", bname))
		return
	}

	tID, ok, err := backend.GetThreadID(board.ID, opID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.Errorf("couldn't find thread: /%s/%d", bname, opID))
		return
	} else if !ok {
		w.WriteHeader(http.StatusNotFound)
		f.FormatError(errors.Errorf("no such thread: /%s/%d", bname, opID))
		return
	}

	post, err := assembleNewPost(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		f.FormatError(err)
		return
	}

	if err := backend.AddReplyToThread(tID, &post); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.Errorf("error: failed to persist reply to /%s/%d", bname, opID))
		return
	}

	thread, err := backend.GetThread(tID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.Errorf("failed to fetch thread: /%s/%d", bname, opID))
		return
	}
	thread.Board = board

	f.FormatThread(thread)
}
