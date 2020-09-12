# TermChan - A simple ~~image~~ board

## What is it?

A simple text board, powered by Golang's stdlib http server and sqlite.

## Installation

Making sure you have the go tool installed and set up, just do

```
go get -u github.com/fgahr/termchan
```

Afterwards, build the project with `go build`, then run with `./termchan`

## Overview

When accessing `/` without any parameters, you will be greeted with a banner and
usage information.

```
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
(*) fields other than content are optional
--------------------------------------------------------------------------------
HAVE FUN!
```

## Setup

```
termchan -h
Usage of termchan:
  -d string
    	the base (configuration) directory for the service (default "./")
  -p int
    	the port for the server to listen on (default 8088)
```
The server will read or create a `config.db` config file and a `boards`
directory to store the database files for the individual boards.

You can add new boards to the `config.db` file with the sqlite3 command and
```
INSERT INTO board (name, description, style)
VALUES ('a', 'a board', 'red');
```
Defining a style is optional but recommended. Recognized style names are, as
of writing, `none`, `black`, `red`, `green`, `yellow`, `blue`, `magenta`,
`cyan`, and `white`.

## Usage

Assuming the server is listening on port 8088 and has a board `/b/`:

```
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

```
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

# TODOs

- Enable banning of users (requires re-enabling tracking of IP addresses, should
  probably mention that in the welcome message)
