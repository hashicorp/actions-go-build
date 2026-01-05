// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

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

func (opts *logOpts) HideFlags() []string {
	return []string{"debug", "v", "q"}
}

func (opts *logOpts) newVerifier(primary, verification build.ResultSource) (*build.Verifier, error) {
	return build.NewVerifier(primary, verification, opts.buildOptions()...)
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

func (opts *logOpts) buildOptions(extraOpts ...build.Option) []build.Option {
	return append([]build.Option{
		build.WithDebugfunc(opts.debugFunc()),
		build.WithLogfunc(opts.logFunc()),
		build.WithLoudfunc(opts.loudFunc()),
	}, extraOpts...)
}
