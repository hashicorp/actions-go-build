package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
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
	if opts.debugFlag || opts.verboseFlag {
		return log.Info
	}
	if opts.quietFlag {
		return log.Discard
	}
	return log.Verbose
}

func (opts *logOpts) loudFunc() log.Func {
	if opts.quietFlag {
		return log.Discard
	}
	return log.Info
}

func (opts *logOpts) buildOptions() []build.Option {
	return []build.Option{
		build.WithLogfunc(opts.logFunc()),
		build.WithDebugfunc(opts.debugFunc()),
	}
}
