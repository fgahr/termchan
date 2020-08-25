package fmt

import (
	"encoding/json"
	"io"

	"github.com/fgahr/termchan/tchan2"
)

type jsonWriter struct {
	enc *json.Encoder
}

func newJSONWriter(w io.Writer) Writer {
	return jsonWriter{json.NewEncoder(w)}
}

func (w jsonWriter) write(obj interface{}) error {
	return w.enc.Encode(obj)
}

func (w jsonWriter) WriteWelcome() error {
	msg := struct {
		Msg string `json:"msg"`
	}{"Welcome to TermChan"}
	return w.enc.Encode(msg)
}

func (w jsonWriter) WriteOverview(boards []tchan2.BoardConfig) error {
	return w.write(boards)
}

func (w jsonWriter) WriteThread(thread tchan2.Thread) error {
	return w.write(thread)
}

func (w jsonWriter) WriteBoard(board tchan2.BoardOverview) error {
	return w.write(board)
}

func (w jsonWriter) WriteError(err error) error {
	wrapper := struct {
		Err string `json:"error"`
	}{err.Error()}
	return w.write(wrapper)
}
