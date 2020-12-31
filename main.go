package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fgahr/termchan/tchan/config"
	"github.com/fgahr/termchan/tchan/http"
	"github.com/fgahr/termchan/tchan/output"
	"github.com/pkg/errors"
)

type command func(conf config.Settings, cmd string, args ...string) error

var commands map[string]command = map[string]command{
	"dump-config":      dumpConfig,
	"create-templates": createTemplates,
	"serve-http":       serveHTTP,
}

func usage(out io.Writer) {
	// NOTE: usage() can not be in the commands slice due to cyclic
	// dependencies during initialization.
	fmt.Fprintf(out, "usage: %s [-d dir] <command>\n", os.Args[0])
	fmt.Fprint(out, `
flags:
  -d <dir>            Set the directory from which to run, defaults to the current directory

commands:
  dump-config         Write the current configuration to stdout; can be used to populate a default config
  create-templates    Place the default templates; will not overwrite existing files
  serve-http          Run as an http service

`)
}

func dumpConfig(conf config.Settings, cmd string, args ...string) error {
	return conf.WriteJSON(os.Stdout)
}

func createTemplates(conf config.Settings, cmd string, args ...string) error {
	log.Println("write templates")
	if err := output.WriteTemplates(conf.TemplateDirectory()); err != nil {
		return errors.Wrapf(err, "%s: creating templates failed", cmd)
	}
	return nil
}

func serveHTTP(conf config.Settings, cmd string, args ...string) error {
	srv, err := http.NewServer(&conf)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT)
	defer close(sigChan)
	go func() {
		for sig := range sigChan {
			log.Printf("caught signal: %v", sig)
			var err error
			switch sig {
			case syscall.SIGHUP:
				err = srv.ReloadConfig()
			case syscall.SIGINT:
				err = srv.Stop()
			default:
				err = errors.Errorf("Unexpected signal: %v", sig)
			}
			if err != nil {
				log.Println(err)
				err = nil
			}
		}
	}()

	return srv.ServeHTTP()
}

func run() error {
	args := os.Args[1:]
	if len(args) == 0 {
		usage(os.Stderr)
		return errors.New("argument required")
	}

	conf := config.Defaults()

	// NOTE: could use the flag package here but right now it wouldn't add much.
	switch arg := args[0]; arg {
	case "-h", "--help", "help":
		usage(os.Stdout)
		return nil
	case "-d":
		if len(args) < 2 {
			usage(os.Stderr)
			return errors.New("-d: directory argument required")
		}
		if err := conf.SetWorkingDirectory(args[1]); err != nil {
			return errors.Wrapf(err, "cannot use working directory %s", args[1])
		}
		args = args[2:]
	}

	if len(args) < 1 {
		usage(os.Stderr)
		return errors.New("command required")
	}

	if err := conf.ReadFromFile(); err != nil {
		return errors.Wrap(err, "failed to read initial configuration")
	}

	cmd := args[0]
	if f, ok := commands[cmd]; ok {
		return f(conf, cmd, args[1:]...)
	}
	usage(os.Stderr)
	return errors.Errorf("no such command: %s", cmd)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
