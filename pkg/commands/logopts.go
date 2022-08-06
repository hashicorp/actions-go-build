package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/log"
)

type logOpts struct {
	debugFlag, verboseFlag, quietFlag bool
}

func (opts *logOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.debugFlag, "debug", false, "debug logging")
	fs.BoolVar(&opts.verboseFlag, "v", false, "verbose logging")
	fs.BoolVar(&opts.quietFlag, "q", false, "quiet logging")
}

func (opts *logOpts) log(f string, a ...any)   { opts.logFunc()(f, a...) }
func (opts *logOpts) debug(f string, a ...any) { opts.debugFunc()(f, a...) }
func (opts *logOpts) loud(f string, a ...any)  { opts.loudFunc()(f, a...) }

func (opts *logOpts) debugFunc() log.Func {
	if opts.debugFlag {
		return log.Info
	}
	if opts.quietFlag {
		return log.Discard
	}
	return log.Debug
}

func (opts *logOpts) logFunc() log.Func {
	if opts.quietFlag {
		return log.Discard
	}
	if opts.debugFlag || opts.verboseFlag {
		return log.Info
	}
	return log.Verbose
}

func (opts *logOpts) loudFunc() log.Func {
	if opts.quietFlag {
		return log.Discard
	}
	return log.Info
}
