package fmt

import (
	"fmt"
	"io"

	"github.com/fgahr/termchan/tchan2"
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

func (w *ansiWriter) WriteWelcome() error {
	for _, line := range bannerTerm {
		w.writeln(fgGreen.FormatANSI(line))
	}
	for _, line := range bannerChan {
		w.writeln(fgBlue.FormatANSI(line))
	}
	w.writeln("Welcome!")
	w.doubleDivider()

	w.hlStyle = fgGreen
	w.writeln(w.hl("Viewing"))
	w.singleDivider()
	w.write("%s the board list\n", w.hl("View"))
	w.write("  curl -s '%s/boards'\n", w.hostname)
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
	w.write("(%s) fields other than content are optional\n", w.hl("*"))
	w.doubleDivider()
	w.write("%s %s!\n", fgGreen.FormatANSI("HAVE"), fgBlue.FormatANSI("FUN"))

	return w.err
}

func (w *ansiWriter) WriteOverview(boards []tchan2.BoardConfig) error {
	return nil
}

func (w *ansiWriter) WriteThread(thread tchan2.Thread) error {
	return nil
}

func (w *ansiWriter) WriteBoard(board tchan2.BoardOverview) error {
	return nil
}

func (w *ansiWriter) WriteError(err error) error {
	return nil
}
