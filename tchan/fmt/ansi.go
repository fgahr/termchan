package fmt

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fgahr/termchan/tchan"
)

type ansiWriter struct {
	hostname string
	out      io.Writer
	hlStyle  Style
	err      error
}

func newANSIWriter(hostname string, w io.Writer) Writer {
	return &ansiWriter{hostname: hostname, out: w}
}

func (w *ansiWriter) write(format string, args ...interface{}) {
	if w.err != nil {
		return
	}

	_, w.err = fmt.Fprintf(w.out, format, args...)
}

func (w *ansiWriter) writeln(args ...interface{}) {
	if w.err != nil {
		return
	}

	_, w.err = fmt.Fprintln(w.out, args...)
}

func (w *ansiWriter) hl(s string) string {
	return w.hlStyle.FormatANSI(s)
}

const (
	singleDiv = "--------------------------------------------------------------------------------"
	doubleDiv = "================================================================================"
)

func (w *ansiWriter) singleDivider() {
	if w.err != nil {
		return
	}

	_, w.err = fmt.Fprintln(w.out, fgBlack.FormatANSI(singleDiv))
}

func (w *ansiWriter) doubleDivider() {
	if w.err != nil {
		return
	}

	_, w.err = fmt.Fprintln(w.out, fgBlack.FormatANSI(doubleDiv))
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

func (w *ansiWriter) WriteWelcome(boards []tchan.BoardConfig) error {
	w.err = nil
	for _, line := range bannerTerm {
		w.writeln(fgGreen.FormatANSI(line))
	}
	for _, line := range bannerChan {
		w.writeln(fgBlue.FormatANSI(line))
	}
	w.writeln("Welcome!")
	w.doubleDivider()
	w.writeln("Boards")
	for _, b := range boards {
		style := getStyle(b.HighlightStyle)
		w.write("  /%s/ - %s\n", style.FormatANSI(b.Name), style.FormatANSI(b.Description))
	}
	w.singleDivider()
	w.writeln("How do I use it?")
	w.singleDivider()

	w.hlStyle = fgGreen
	w.writeln(w.hl("Viewing"))
	w.singleDivider()
	w.write("%s a board (e.g. /g/)\n", w.hl("View"))
	w.write("  curl -s '%s/g'\n", w.hostname)
	w.singleDivider()
	w.write("%s a thread (e.g. thread #23 on /v/)\n", w.hl("View"))
	w.write("  curl -s '%s/v/23'\n", w.hostname)
	w.singleDivider()
	w.write("%s as JSON\n", w.hl("View"))
	w.write("  curl -s '%s/d/69?format=json'\n", w.hostname)
	w.doubleDivider()

	w.hlStyle = fgBlue
	w.writeln(w.hl("Posting"))
	w.singleDivider()
	w.write("%s a reply to a thread (%s)\n", w.hl("Post"), w.hl("*"))
	w.write("  curl -s '%s/g/42' \\\n", w.hostname)
	w.write("      --data-urlencode \"format=json\" \\\n")
	w.write("      --data-urlencode \"name=ilovebsd\" \\\n")
	w.write("      --data-urlencode \"content=Have you considered OpenBSD?\"\n")
	w.singleDivider()
	w.write("%s (i.e. create) a thread (%s)\n", w.hl("Post"), w.hl("*"))
	w.write("  curl -s '%s/b' \\\n", w.hostname)
	w.write("      --data-urlencode \"name=m00t\" \\\n")
	w.write("      --data-urlencode \"topic=Candlejack\" \\\n")
	w.write("      --data-urlencode \"content=I'm not afraid of him, what's he gon-\"\n")
	w.singleDivider()
	w.write("(%s) fields other than content are optional, board/thread has to exist\n", w.hl("*"))
	w.doubleDivider()

	w.write("%s %s!\n", fgGreen.FormatANSI("HAVE"), fgBlue.FormatANSI("FUN"))

	return w.err
}

func (w *ansiWriter) writePost(p tchan.Post) {
	if w.err != nil {
		return
	}

	w.write("[%d] %s wrote at %s\n", p.ID, p.Author, p.Timestamp.Format(time.ANSIC))
	w.writeln()
	for _, line := range strings.Split(p.Content, "\n") {
		w.writeln(line)
	}
}

func (w *ansiWriter) WriteThread(t tchan.Thread) error {
	w.err = nil
	w.hlStyle = getStyle(t.Board.HighlightStyle)

	w.write("/%s/%d %s\n", w.hl(t.Board.Name), t.Posts[0].ID, w.hl(t.Topic))
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
	w.write("%d %s\n", numReplies, rep)
	return w.err
}

func (w *ansiWriter) writeThreadOverview(t tchan.ThreadOverview, bc tchan.BoardConfig) {
	if w.err != nil {
		return
	}

	rep := "replies"
	if t.NumReplies == 1 {
		rep = "reply"
	}
	w.write("/%s/%d %s (%d %s) updated %s\n",
		w.hl(bc.Name), t.OP.ID, w.hl(t.Topic),
		t.NumReplies, rep, t.Active.Format(time.ANSIC))
	w.singleDivider()
	w.writePost(t.OP)
}

func (w *ansiWriter) WriteBoard(board tchan.BoardOverview) error {
	w.err = nil
	bc := board.MetaData
	w.hlStyle = getStyle(bc.HighlightStyle)
	w.write("/%s/ - %s\n", w.hl(board.MetaData.Name), w.hl(board.MetaData.Description))

	for _, thread := range board.Threads {
		w.doubleDivider()
		w.writeThreadOverview(thread, bc)
	}

	w.doubleDivider()
	thr := "threads"
	if len(board.Threads) == 1 {
		thr = "thread"
	}
	w.write("%d %s\n", len(board.Threads), thr)

	return w.err
}

func (w *ansiWriter) WriteError(err error) error {
	w.write("%s: %s\n", fgRed.FormatANSI("ERROR"), err.Error())
	return w.err
}
