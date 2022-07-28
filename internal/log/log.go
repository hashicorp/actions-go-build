package log

import (
	"log"
	"os"

	"golang.org/x/term"
)

type logFunc func(string, ...any)

var Info, Verbose = func() (info, verbose logFunc) {
	l := log.New(os.Stderr, "", 0)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		l.SetFlags(log.LstdFlags)
		return l.Printf, l.Printf
	}
	return l.Printf, func(string, ...any) {}
}()
