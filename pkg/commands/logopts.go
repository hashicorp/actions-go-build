package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/log"
)

type logOpts struct {
	debug, verbose, quiet bool
}

func (opts *logOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.debug, "debug", false, "debug logging")
	fs.BoolVar(&opts.verbose, "v", false, "verbose logging")
	fs.BoolVar(&opts.quiet, "q", false, "quiet logging")
}

func (opts *logOpts) debugFunc() log.Func {
	if opts.debug {
		return log.Info
	}
	if opts.quiet {
		return log.Discard
	}
	return log.Debug
}

func (opts *logOpts) logFunc() log.Func {
	if opts.quiet {
		return log.Discard
	}
	if opts.verbose {
		return log.Info
	}
	return log.Verbose
}
