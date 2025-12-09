// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package log

import (
	"log"
	"os"
	"strings"

	"golang.org/x/term"
)

type Func func(string, ...any)

// IsTerm returns true if we appear to be connected to a tty,
// indicating that there is likely to be a user present.
// Usually in that case we should default to quieter logging
// than we would in an unattended scanario, where greater output
// volume is more easily tolerated and useful.
func IsTerm() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func IsDebug() bool {
	switch strings.ToLower(os.Getenv("DEBUG")) {
	default:
		return false
	case "true", "1", "y", "yes":
		return true
	}
}

func IsVerbose() bool {
	return !IsTerm() || IsDebug()
}

func IsInfo() bool {
	return true
}

var nothing = func(string, ...any) {}

var Info, Verbose, Debug, Discard = func() (info, verbose, debug, discard Func) {
	info, verbose, debug, discard = nothing, nothing, nothing, nothing
	l := log.New(os.Stderr, "", 0)
	if !IsTerm() {
		l.SetFlags(log.LstdFlags)
	}
	if IsInfo() {
		info = l.Printf
	}
	if IsVerbose() {
		verbose = l.Printf
	}
	if IsDebug() {
		debug = func(f string, a ...any) {
			l.Printf("DEBUG: "+f, a...)
		}
	}
	return
}()
