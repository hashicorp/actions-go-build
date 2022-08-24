package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/hashicorp/actions-go-build/internal/log"
)

// Settings contains settings for running builds.
// These are not to be confused with build.Config, these settings
// are build-run specific and not part of the _definition_ of the build.
// Don't use this directly, use the With... functions to set
// settings when calling New.
type Settings struct {
	bash         string
	context      context.Context
	Log          func(string, ...any)
	Debug        func(string, ...any)
	Loud         func(string, ...any)
	stdout       io.Writer
	stderr       io.Writer
	forceRebuild bool
	cleanOnly    bool
	logPrefix    string
}

// Option represents a function that configures Settings.
type Option func(*Settings)

// WithContext sets the context passed when we shell out.
func WithContext(c context.Context) Option { return func(s *Settings) { s.context = c } }

// WithLogfunc sets the log func.
func WithLogfunc(f func(string, ...any)) Option {
	return func(s *Settings) { s.Log = s.makeLogFunc(f) }
}

// WithLogPrefix sets the log prefix.
func WithLogPrefix(p string) Option { return func(s *Settings) { s.logPrefix = p } }

// WithDebugfunc sest the debug func.
func WithDebugfunc(f func(string, ...any)) Option {
	return func(s *Settings) { s.Debug = s.makeLogFunc(f) }
}

// WithDebugfunc sest the debug func.
func WithLoudfunc(f func(string, ...any)) Option {
	return func(s *Settings) { s.Loud = s.makeLogFunc(f) }
}

// WithStdout sets the stdout for when we shell out.
func WithStdout(w io.Writer) Option { return func(s *Settings) { s.stdout = w } }

// WithStderr sets the stderr for when we shell out.
func WithStderr(w io.Writer) Option { return func(s *Settings) { s.stderr = w } }

// WithForceRebuild forces a build to be re-done rather than using cache.
func WithForceRebuild(on bool) Option { return func(s *Settings) { s.forceRebuild = on } }

func WithCleanOnly(on bool) Option { return func(s *Settings) { s.cleanOnly = on } }

func newSettings(options []Option) (Settings, error) {
	s := &Settings{}
	err := s.setOptions(options...)
	return *s, err
}

func (s *Settings) setOptions(opts ...Option) error {
	for _, o := range opts {
		o(s)
	}
	return s.setDefaults()
}

func (s *Settings) setDefaults() (err error) {
	s.bash, err = resolveBashPath(s.bash)
	if err != nil {
		return err
	}
	if s.context == nil {
		s.context = context.Background()
	}
	if s.Debug == nil {
		s.Debug = log.Debug
	}
	if s.Log == nil {
		s.Log = log.Verbose
	}
	if s.Loud == nil {
		s.Loud = log.Info
	}

	WithDebugfunc(s.Debug)(s)
	WithLogfunc(s.Log)(s)
	WithLoudfunc(s.Loud)(s)
	if s.stdout == nil {
		s.stdout = os.Stderr
	}
	if s.stderr == nil {
		s.stderr = os.Stderr
	}
	return nil
}

func resolveBashPath(path string) (string, error) {
	if path == "" {
		path = "bash"
	}
	return exec.LookPath(path)
}

func (s *Settings) makeLogFunc(logFunc log.Func) log.Func {
	return makePrefixedLocFunc(s.logPrefix, logFunc)
}

func makePrefixedLocFunc(name string, logFunc log.Func) log.Func {
	if name != "" {
		name = name + ": "
	}
	return func(f string, a ...any) {
		f = fmt.Sprintf("%s%s", name, f)
		logFunc(f, a...)
	}
}
