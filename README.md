# TermChan - A simple ~~image~~ board

## What is it?

A simple text board, powered by Golang's stdlib http server and sqlite.

## Running and Configuring

Making sure you have the go tool installed and set up, just do

```sh
go get -u github.com/fgahr/termchan
```

Afterwards, build the project with `go build`, then run it

```sh
$ termchan help
usage: ./termchan [-d dir] <command>

flags:
  -d <dir>            Set the directory from which to run, defaults to the current directory

commands:
  dump-config         Write the current configuration to stdout; can be used to populate a default config
  create-templates    Place the default templates; will not overwrite existing files
  serve-http          Run as an http service

```

The server will read a `config.json` file, a template of which can be created via

```sh
$ termchan dump-config > config.json
```

The list of boards can be adjusted in `config.json`, using the example board
as a template. After changing the configuration, restart the server or send a
SIGHUP signal, e.g.

```sh
$ kill -s HUP $(pgrep termchan)
```

to make it reload its config.

## Overview

When accessing `/` without any parameters, you will be greeted with a banner and
usage information. An example of HTML output for the banner is
[here](welcome.html).

```sh
$ curl -s 'localhost:8088/'
  ::::::::::::.,:::::: :::::::..   .        :
  ;;;;;;;;'''';;;;'''' ;;;;``;;;;  ;;,.    ;;;
       [[      [[cccc   [[[,/[[['  [[[[, ,[[[[,
       $$      $$""""   $$$$$$c    $$$$$$$$"$$$
       88,     888oo,__ 888b "88bo,888 Y88" 888o
       MMM     """"YUMMMMMMM   "W" MMM  M'  "MMM
                                      .,-:::::   ::   .:   :::.   :::.    :::.
                                    ,;;;'````'  ,;;   ;;,  ;;`;;  `;;;;,  `;;;
                                    [[[        ,[[[,,,[[[ ,[[ '[[,  [[[[[. '[[
                                    $$$        "$$$"""$$$c$$$cc$$$c $$$ "Y$c$$
                                    `88bo,__,o, 888   "88o888   888,888    Y88
                                      "YUMMMMMP"MMM    YMMYMM   ""` MMM     YM
Welcome!
================================================================================
Boards
  /b/ - Random
  /g/ - Technology
  /m/ - Meta
  /v/ - Games
--------------------------------------------------------------------------------
How do I use it?
--------------------------------------------------------------------------------
Viewing
================================================================================
View a board (e.g. /g/)
  curl -s 'localhost:8088/g'
--------------------------------------------------------------------------------
View a board as HTML (e.g. /m/)
  curl -s 'localhost:8088/m?format=html'
--------------------------------------------------------------------------------
View a thread (e.g. thread #23 on /v/)
  curl -s 'localhost:8088/v/23'
--------------------------------------------------------------------------------
View a thread as JSON
  curl -s 'localhost:8088/d/69?format=json'
--------------------------------------------------------------------------------
Posting
================================================================================
Post a reply to a thread (*)
  curl -s 'localhost:8088/g/42' \
      --data-urlencode "format=json" \
      --data-urlencode "name=ilovebsd" \
      --data-urlencode "content=Have you considered OpenBSD?"
--------------------------------------------------------------------------------
Post (i.e. create) a thread (*)
  curl -s 'localhost:8088/b' \
      --data-urlencode "name=m00t" \
      --data-urlencode "topic=Candlejack" \
      --data-urlencode "content=I'm not afraid of him, what's he gon-"
--------------------------------------------------------------------------------
(*) fields other than content are optional, board/thread has to exist
================================================================================
HAVE FUN!
```

## Usage

### With tccli

There is the [tccli](https://github.com/fgahr/termchan-cli) tool to simplify
common operations without needing to interact with `curl` directly.

### With curl

Assuming the server is listening on port 8088 and has a board `/b/`, post with

```sh
$ curl -s 'localhost:8088/b'  \
      --data-urlencode "name=me" \
      --data-urlencode "content=This is my first post"
foo
================================================================================
[1] me wrote at Sat Jul 11 10:37:07 2020

This is my first post
================================================================================
0 replies
```

You will be greeted with a JSON view of the newly created thread. You can get an
overview of the board with

```sh
$ curl -s 'localhost:8088/b?format=json' | jq
{
  "name": "b",
  "description": "Random",
  "threads": [
    {
      "topic": "foo",
      "op": {
        "id": 1,
        "author": "me",
        "content": "This is my first post",
        "time": "2020-07-11T10:27:46Z"
      },
      "replies": 0,
      "postedAt": "2020-07-11T10:27:46Z",
      "latestReplyAt": "2020-07-11T10:27:46Z"
    }
  ]
}
```

## Advanced

### Appearance

The appearance of termchan can be changed via [templates](https://golang.org/pkg/text/template/). These are part of Go's
standard library. To support both terminal and html output, both `text/template`
and `html/template` are used. The packages `tchan/output/ansi` and
`tchan/output/html` are mostly mirrored and either one is suitable to learn
about available fields and functions from within the template. Running

```sh
$ termchan create-templates
```

will dump the integrated defaults as files inside a `template/` folder. Already
existing files will not be overwritten. If you delete a template, its default
will be used when running termchan.

### Domain Socket Connections

In the `config.json` file, the default transport type is `tcp` on `:8088`.
However, for reverse proxy setups, connection via a domain socket can be used.
E.g.
```
...
	"transport": {
		"Protocol": "unix",
		"Socket": "/tmp/termchan/socket"
	},
...
```

## TODOs

- Enable banning of users (requires re-enabling tracking of IP addresses, should
  probably mention that in the welcome message)
- Basic security measures
- More available styles (e.g. bold)
- Enable editing CSS for html output
- Whatever reasonable request you might open an issue for (pull-requests welcome)
