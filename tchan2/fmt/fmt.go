package fmt

import (
	"io"
	"net/http"
	"net/url"

	"github.com/fgahr/termchan/tchan2"
)

// Writer describes an entity in charge of writing a server response.
type Writer interface {
	WriteWelcome(boards []tchan2.BoardConfig) error
	WriteThread(thread tchan2.Thread) error
	WriteBoard(board tchan2.BoardOverview) error
	WriteError(err error) error
}

// GetWriter finds a suitable writer for the request.
func GetWriter(params url.Values, r *http.Request, w io.Writer) Writer {
	switch params.Get("format") {
	case "json":
		return newJSONWriter(w)
	default:
		return newANSIWriter(r.Host, w)
	}
}
