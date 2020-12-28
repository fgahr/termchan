package json

import (
	"encoding/json"
	"net/http"

	"github.com/fgahr/termchan/tchan"
)

type Writer struct {
	res http.ResponseWriter
	enc *json.Encoder
}

func NewWriter(r *http.Request, w http.ResponseWriter) *Writer {
	return &Writer{res: w, enc: json.NewEncoder(w)}
}

func (w *Writer) write(obj interface{}) error {
	return w.enc.Encode(obj)
}

func (w *Writer) WriteWelcome(boards []tchan.BoardConfig) error {
	return w.write(boards)
}

func (w *Writer) WriteThread(thread tchan.Thread) error {
	return w.write(thread)
}

func (w *Writer) WriteBoard(board tchan.BoardOverview) error {
	return w.write(board)
}

func (w *Writer) WriteError(status int, err error) error {
	wrapper := struct {
		Status int    `json:"status"`
		Error  string `json:"error"`
	}{Status: status, Error: err.Error()}
	w.res.WriteHeader(status)
	return w.write(wrapper)
}
