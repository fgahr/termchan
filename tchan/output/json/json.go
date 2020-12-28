package json

import (
	"encoding/json"
	"io"

	"github.com/fgahr/termchan/tchan"
)

type jsonWriter struct {
	enc *json.Encoder
}

func newJSONWriter(w io.Writer) interface{} {
	return jsonWriter{json.NewEncoder(w)}
}

func (w jsonWriter) write(obj interface{}) error {
	return w.enc.Encode(obj)
}

func (w jsonWriter) WriteWelcome(boards []tchan.BoardConfig) error {
	return w.write(boards)
}

func (w jsonWriter) WriteThread(thread tchan.Thread) error {
	return w.write(thread)
}

func (w jsonWriter) WriteBoard(board tchan.BoardOverview) error {
	return w.write(board)
}

func (w jsonWriter) WriteError(status int, err error) error {
	wrapper := struct {
		Status int    `json:"status"`
		Error  string `json:"error"`
	}{Status: status, Error: err.Error()}
	return w.write(wrapper)
}
