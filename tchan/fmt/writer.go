package fmt

import (
	"io"
	"strings"
	"time"

	"github.com/fgahr/termchan/tchan"
)

type presenter interface {
	clean(s string) string
	apply(sty Style, s string) string
	write(format string, args ...interface{}) error
	header() error
	footer() error
}

type writer struct {
	pres     presenter
	hostname string
	hlStyle  Style
	err      error
}

func newANSIWriter(hostname string, w io.Writer) Writer {
	return &writer{hostname: hostname, pres: ansiPresenter{w}}
}

func newHTMLWriter(hostname string, w io.Writer) Writer {
	return &writer{hostname: hostname, pres: htmlPresenter{w}}
}

func (w *writer) write(format string, args ...interface{}) {
	if w.err != nil {
		return
	}

	w.err = w.pres.write(format, args...)
}

func (w *writer) hl(s string) string {
	return w.pres.apply(w.hlStyle, s)
}

func (w *writer) apply(sty Style, s string) string {
	return w.pres.apply(sty, s)
}

const (
	singleDiv = "--------------------------------------------------------------------------------"
	doubleDiv = "================================================================================"
)

func (w *writer) singleDivider() {
	w.write(w.apply(FgBlack, singleDiv))
}

func (w *writer) doubleDivider() {
	w.write(w.apply(FgBlack, doubleDiv))
}

// Created with figlet's cosmic.flf font
var bannerTerm = []string{
	"  ::::::::::::.,:::::: :::::::..   .        :",
	"  ;;;;;;;;'''';;;;'''' ;;;;``;;;;  ;;,.    ;;;",
	"       [[      [[cccc   [[[,/[[['  [[[[, ,[[[[,",
	"       $$      $$\"\"\"\"   $$$$$$c    $$$$$$$$\"$$$",
	"       88,     888oo,__ 888b \"88bo,888 Y88\" 888o",
	"       MMM     \"\"\"\"YUMMMMMMM   \"W\" MMM  M'  \"MMM",
}
var bannerChan = []string{
	"                                      .,-:::::   ::   .:   :::.   :::.    :::.",
	"                                    ,;;;'````'  ,;;   ;;,  ;;`;;  `;;;;,  `;;;",
	"                                    [[[        ,[[[,,,[[[ ,[[ '[[,  [[[[[. '[[",
	"                                    $$$        \"$$$\"\"\"$$$c$$$cc$$$c $$$ \"Y$c$$",
	"                                    `88bo,__,o, 888   \"88o888   888,888    Y88",
	"                                      \"YUMMMMMP\"MMM    YMMYMM   \"\"` MMM     YM",
}

func (w *writer) WriteWelcome(boards []tchan.BoardConfig) error {
	w.err = w.pres.header()
	defer w.pres.footer()

	for _, line := range bannerTerm {
		w.write(w.apply(FgGreen, line))
	}
	for _, line := range bannerChan {
		w.write(w.apply(FgBlue, line))
	}
	w.write("Welcome!")
	w.doubleDivider()
	w.write("Boards")
	for _, b := range boards {
		sty := getStyle(b.HighlightStyle)
		w.write("  /%s/ - %s", w.apply(sty, b.Name), w.apply(sty, b.Description))
	}
	w.singleDivider()
	w.write("How do I use it?")
	w.singleDivider()

	w.hlStyle = FgGreen
	w.write(w.hl("Viewing"))
	w.singleDivider()
	w.write("%s a board (e.g. /g/)", w.hl("View"))
	w.write("  curl -s '%s/g'", w.hostname)
	w.singleDivider()
	w.write("%s a board as HTML (e.g. /m/)", w.hl("View"))
	w.write("  curl -s '%s/m?format=html'", w.hostname)
	w.singleDivider()
	w.write("%s a thread (e.g. thread #23 on /v/)", w.hl("View"))
	w.write("  curl -s '%s/v/23'", w.hostname)
	w.singleDivider()
	w.write("%s as JSON", w.hl("View"))
	w.write("  curl -s '%s/d/69?format=json'", w.hostname)
	w.doubleDivider()

	w.hlStyle = FgBlue
	w.write(w.hl("Posting"))
	w.singleDivider()
	w.write("%s a reply to a thread (%s)", w.hl("Post"), w.hl("*"))
	w.write("  curl -s '%s/g/42' \\", w.hostname)
	w.write("      --data-urlencode \"format=json\" \\")
	w.write("      --data-urlencode \"name=ilovebsd\" \\")
	w.write("      --data-urlencode \"content=Have you considered OpenBSD?\"")
	w.singleDivider()
	w.write("%s (i.e. create) a thread (%s)", w.hl("Post"), w.hl("*"))
	w.write("  curl -s '%s/b' \\", w.hostname)
	w.write("      --data-urlencode \"name=m00t\" \\")
	w.write("      --data-urlencode \"topic=Candlejack\" \\")
	w.write("      --data-urlencode \"content=I'm not afraid of him, what's he gon-\"")
	w.singleDivider()
	w.write("(%s) fields other than content are optional, board/thread has to exist", w.hl("*"))
	w.doubleDivider()

	w.write("%s %s!", w.apply(FgGreen, "HAVE"), w.apply(FgBlue, "FUN"))

	return w.err
}

func (w *writer) writePost(p tchan.Post) {
	if w.err != nil {
		return
	}

	w.write("[%d] %s wrote at %s", p.ID, p.Author, p.Timestamp.Format(time.ANSIC))
	w.write("")
	for _, line := range strings.Split(p.Content, "\n") {
		w.write(line)
	}
}

func (w *writer) WriteThread(t tchan.Thread) error {
	w.err = w.pres.header()
	defer w.pres.footer()

	w.hlStyle = getStyle(t.Board.HighlightStyle)

	w.write("/%s/%d %s", w.hl(t.Board.Name), t.Posts[0].ID, w.hl(t.Topic))
	w.doubleDivider()

	for i, post := range t.Posts {
		if i > 0 {
			w.singleDivider()
		}
		w.writePost(post)
	}

	w.doubleDivider()
	numReplies := len(t.Posts) - 1
	rep := "replies"
	if numReplies == 1 {
		rep = "reply"
	}
	w.write("%d %s", numReplies, rep)
	return w.err
}

func (w *writer) writeThreadOverview(t tchan.ThreadOverview, bc tchan.BoardConfig) {
	if w.err != nil {
		return
	}

	rep := "replies"
	if t.NumReplies == 1 {
		rep = "reply"
	}
	w.write("/%s/%d %s (%d %s) updated %s",
		w.hl(bc.Name), t.OP.ID, w.hl(t.Topic),
		t.NumReplies, rep, t.Active.Format(time.ANSIC))
	w.singleDivider()
	w.writePost(t.OP)
}

func (w *writer) WriteBoard(board tchan.BoardOverview) error {
	w.err = w.pres.header()
	defer w.pres.footer()

	bc := board.MetaData
	w.hlStyle = getStyle(bc.HighlightStyle)
	w.write("/%s/ - %s", w.hl(board.MetaData.Name), w.hl(board.MetaData.Description))

	for _, thread := range board.Threads {
		w.doubleDivider()
		w.writeThreadOverview(thread, bc)
	}

	w.doubleDivider()
	thr := "threads"
	if len(board.Threads) == 1 {
		thr = "thread"
	}
	w.write("%d %s", len(board.Threads), thr)

	return w.err
}

func (w *writer) WriteError(err error) error {
	w.err = w.pres.header()
	defer w.pres.footer()

	w.write("%s: %s", w.apply(FgRed, "ERROR"), err.Error())
	return w.err
}
