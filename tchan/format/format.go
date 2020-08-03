package format

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/fgahr/termchan/tchan/data"
	"github.com/fgahr/termchan/tchan/format/ansi"
)

// Formatter describes an entity presenting output in response to a request.
type Formatter interface {
	FormatWelcome(params *data.BoardParameters, hostname string)
	FormatThread(t *data.Thread)
	FormatBoard(b *data.Board)
	FormatError(err error)
}

type jsonFormatter struct {
	w   http.ResponseWriter
	enc *json.Encoder
	err error
}

func newJSONFormatter(w http.ResponseWriter) *jsonFormatter {
	enc := json.NewEncoder(w)
	return &jsonFormatter{w, enc, nil}
}

func (f *jsonFormatter) write(obj interface{}) {
	if f.err != nil {
		return
	}

	f.err = f.enc.Encode(obj)
	if f.err != nil {
		f.w.WriteHeader(http.StatusInternalServerError)
		log.Println(f.err)
	}
}

func (f *jsonFormatter) FormatWelcome(params *data.BoardParameters, hostname string) {
	f.write(params)
}

func (f *jsonFormatter) FormatThread(t *data.Thread) {
	f.write(t)
}

func (f *jsonFormatter) FormatBoard(b *data.Board) {
	f.write(b)
}

// WrappedError wraps an error message to be formatted as JSON.
type WrappedError struct {
	Err string `json:"error"`
}

func (f jsonFormatter) FormatError(err error) {
	f.write(err)
}

type terminalFormatter struct {
	w       http.ResponseWriter
	hlStyle ansi.Style // foreground style
	err     error
}

func newTerminalFormatter(w http.ResponseWriter) *terminalFormatter {
	return &terminalFormatter{w, ansi.FgWhite, nil}
}

func (f *terminalFormatter) write(format string, args ...interface{}) {
	if f.err != nil {
		return
	}

	_, f.err = fmt.Fprintf(f.w, format, args...)
	if f.err != nil {
		f.w.WriteHeader(http.StatusInternalServerError)
		log.Println(f.err)
	}
}

func (f *terminalFormatter) hl(v interface{}) string {
	return fmt.Sprintf("%s%v%s", f.hlStyle, v, ansi.Reset)
}

func hl(style ansi.Style, v interface{}) string {
	return fmt.Sprintf("%s%v%s", style, v, ansi.Reset)
}

func (f *terminalFormatter) insertDivider(symbol byte) {
	if f.err != nil {
		return
	}

	f.write("%s", ansi.FgBlack)
	for i := 0; i < 80; i++ {
		f.write("%c", symbol)
	}
	f.write("%s\n", ansi.Reset)
}

const (
	// Banner color for TERM
	tc = ansi.FgGreen
	// Banner color for CHAN
	cc = ansi.FgBlue
)

var terminalBanner = []string{
	// Created with figlet's cosmic.flf font
	tc + "  ::::::::::::.,:::::: :::::::..   .        :",
	tc + "  ;;;;;;;;'''';;;;'''' ;;;;``;;;;  ;;,.    ;;;",
	tc + "       [[      [[cccc   [[[,/[[['  [[[[, ,[[[[,",
	tc + "       $$      $$\"\"\"\"   $$$$$$c    $$$$$$$$\"$$$",
	tc + "       88,     888oo,__ 888b \"88bo,888 Y88\" 888o",
	tc + "       MMM     \"\"\"\"YUMMMMMMM   \"W\" MMM  M'  \"MMM",
	cc + "                                      .,-:::::   ::   .:   :::.   :::.    :::.",
	cc + "                                    ,;;;'````'  ,;;   ;;,  ;;`;;  `;;;;,  `;;;",
	cc + "                                    [[[        ,[[[,,,[[[ ,[[ '[[,  [[[[[. '[[",
	cc + "                                    $$$        \"$$$\"\"\"$$$c$$$cc$$$c $$$ \"Y$c$$",
	cc + "                                    `88bo,__,o, 888   \"88o888   888,888    Y88",
	cc + "                                      \"YUMMMMMP\"MMM    YMMYMM   \"\"` MMM     YM",
}

func (f *terminalFormatter) FormatWelcome(params *data.BoardParameters, hostname string) {
	for _, line := range terminalBanner {
		f.write("%s%s\n", line, ansi.Reset)
	}

	f.write("Welcome!\n")
	f.insertDivider('=')
	f.write("Boards\n")
	for _, b := range params.Boards {
		color := b.HighlightColor
		f.write("    /%s/ - %s\n", hl(color, b.Name), hl(color, b.Description))
	}
	f.insertDivider('-')

	f.hlStyle = ansi.FgGreen
	f.write("How do I use it?\n")
	f.insertDivider('-')
	f.write("%s\n", f.hl("Viewing"))
	f.insertDivider('=')
	f.write("%s a board (e.g. /g/):\n", f.hl("View"))
	f.write("  curl -s '%s/g'\n", hostname)
	f.insertDivider('-')
	f.write("%s a thread (e.g. thread #23 on /v/):\n", f.hl("View"))
	f.write("  curl -s '%s/v/23'\n", hostname)
	f.insertDivider('-')
	f.write("%s a thread as JSON:\n", f.hl("View"))
	f.write("  curl -s '%s/d/69?format=json'\n", hostname)
	f.insertDivider('-')

	f.hlStyle = ansi.FgBlue
	f.write("%s\n", f.hl("Posting"))
	f.insertDivider('=')
	f.write("%s a reply to a thread:\n", f.hl("Post"))
	f.write("  curl -s '%s/g/42' \\\n", hostname)
	f.write("      --data-urlencode \"format=json\" \\\n")
	f.write("      --data-urlencode \"name=ilovebsd\" \\\n")
	f.write("      --data-urlencode \"content=Have you considered OpenBSD?\"\n")
	f.insertDivider('-')
	f.write("%s (i.e. create) a thread:\n", f.hl("Post"))
	f.write("  curl -s '%s/b' \\\n", hostname)
	f.write("      --data-urlencode \"name=m00t\" \\\n")
	f.write("      --data-urlencode \"topic=Candlejack\" \\\n")
	f.write("      --data-urlencode \"content=I'm not afraid of him. What's he gon-\"\n")
	f.insertDivider('-')
	f.write("%s: only the content field is mandatory\n", f.hl("NOTE"))
	f.insertDivider('-')

	// f.write("Usage (* = HOST:PORT)\n")
	// f.hlStyle = ansi.FgGreen
	// f.write("    curl */b                                     (%s) board view\n", f.hl("GET"))
	// f.write("    curl */b/1 -G --data-urlencode \"format=json\" (%s) thread view, JSON output\n", f.hl("GET"))
	// f.write("    curl */b --data-urlencode \"topic=foo\"        (%s) create thread\n", f.hl("POST"))
	// f.write("    curl */b/1 --data \"bar\"         (%s) reply to thread\n", f.hl("POST"))
	// f.insertDivider('-')

	// f.write("Parameters (* = mandatory, all others optional)\n")
	// f.hlStyle = ansi.FgBlue
	// f.write("    format=json                 (%s/%s) request JSON output\n", f.hl("GET"), f.hl("POST"))
	// f.write("    name=m00t                       (%s) name when posting\n", f.hl("POST"))
	// f.write("    topic=The%%20Game                (%s) topic when creating a thread\n", f.hl("POST"))
	// f.write("    content=Hello%%2C%%20World%%21     (%s) content of a new thread or a reply\n", f.hl("POST"))
	// f.insertDivider('-')

	// f.write("Limits\n")
	// f.write("    Post size (in bytes):          %6d\n", h.PostSize)
	// f.write("    Thread count (per board):      %6d\n", h.ThreadLimit)
	// f.write("    Reply count (per thread):      %6d\n", h.ReplyLimit)
	// f.insertDivider('=')

	f.write("%s %s!\n", hl(ansi.FgGreen, "HAVE"), hl(ansi.FgBlue, "FUN"))
}

func (f *terminalFormatter) writePost(p data.Post) {
	f.write("[%s] %s wrote at %s\n",
		f.hl(p.InBoardID), f.hl(p.Author), p.Timestamp.Format(time.ANSIC))
	f.write("\n")
	f.write("%s\n", p.Content)
}

func (f *terminalFormatter) writeThreadOverview(t *data.ThreadOverview) {
	if f.err != nil {
		return
	}

	rep := "replies"
	if t.ReplyCount == 1 {
		rep = "reply"
	}
	f.write("/%s/%d %s (%d %s) updated %s\n",
		f.hl(t.Board.Name), t.OP.InBoardID, f.hl(t.Topic),
		t.ReplyCount, rep, t.LastReply.Format(time.ANSIC))
	f.insertDivider('-')
	f.writePost(t.OP)
}

func (f *terminalFormatter) FormatBoard(b *data.Board) {
	f.hlStyle = b.HighlightStyle
	f.write("/%s/ - %s\n", f.hl(b.Name), f.hl(b.Description))

	for _, thread := range b.ActiveThreads {
		f.insertDivider('=')
		f.writeThreadOverview(&thread)
	}

	f.insertDivider('=')
	thr := "threads"
	if len(b.ActiveThreads) == 1 {
		thr = "thread"
	}
	f.write("%d %s\n", len(b.ActiveThreads), thr)
}

func (f *terminalFormatter) FormatThread(t *data.Thread) {
	f.hlStyle = t.Board.HighlightColor
	if len(t.Posts) == 0 {
		f.w.WriteHeader(http.StatusInternalServerError)
		f.FormatError(errors.New("no posts found"))
		log.Printf("no posts in thread '%s'", t.Topic)
		return
	}

	f.write("/%s/%d %s\n", f.hl(t.Board.Name), t.OP().InBoardID, f.hl(t.Topic))
	f.insertDivider('=')

	for i, post := range t.Posts {
		if i > 0 {
			f.insertDivider('-')
		}
		f.writePost(post)
	}

	f.insertDivider('=')
	numReplies := len(t.Posts) - 1
	rep := "replies"
	if numReplies == 1 {
		rep = "reply"
	}
	f.write("%d %s\n", numReplies, rep)
}

func (f *terminalFormatter) FormatError(err error) {
	f.write("%sERROR:%s %s\n", ansi.FgRed, ansi.Reset, err)
}

// Select determines and fetches the appropriate output formatter for a request.
func Select(query url.Values, w http.ResponseWriter) Formatter {
	format := query.Get("format")
	switch format {
	case "json":
		return newJSONFormatter(w)
	default:
		return newTerminalFormatter(w)
	}
}
