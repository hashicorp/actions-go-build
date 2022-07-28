package log

import (
	"log"
	"os"

	"golang.org/x/term"
)

type logFunc func(string, ...any)

// IsTerm returns true if we appear to be connected to a tty,
// indicating that there is likely to be a user present.
// Usually in that case we should default to quieter logging
// than we would in an unattended scanario, where greater output
// volume is more easily tolerated and useful.
func IsTerm() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

var Info, Verbose = func() (info, verbose logFunc) {
	l := log.New(os.Stderr, "", 0)
	if !IsTerm() {
		l.SetFlags(log.LstdFlags)
		return l.Printf, l.Printf
	}
	return l.Printf, func(string, ...any) {}
}()
