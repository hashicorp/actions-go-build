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
	name         string
	context      context.Context
	Log          func(string, ...any)
	Debug        func(string, ...any)
	Loud         func(string, ...any)
	stdout       io.Writer
	stderr       io.Writer
	forceRebuild bool
}

func makeNamedLogFunc(name string, logFunc log.Func) log.Func {
	return func(f string, a ...any) {
		f = fmt.Sprintf("%s: %s", name, f)
		logFunc(f, a...)
	}
}

func (s *Settings) makeLogFunc(logFunc log.Func) log.Func {
	return makeNamedLogFunc(s.name, logFunc)
}

// Option represents a function that configures Settings.
type Option func(*Settings)

// WithContext sets the context passed when we shell out.
func WithContext(c context.Context) Option { return func(s *Settings) { s.context = c } }

// WithLogfunc sets the log func.
func WithLogfunc(f func(string, ...any)) Option {
	return func(s *Settings) { s.Log = s.makeLogFunc(f) }
}

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

func newSettings(name string, options []Option) (Settings, error) {
	if name == "" {
		name = "unnamed"
	}
	out := &Settings{
		name: name,
	}
	for _, o := range options {
		o(out)
	}
	if err := out.setDefaults(); err != nil {
		return Settings{}, err
	}
	return *out, nil
}

func resolveBashPath(path string) (string, error) {
	if path == "" {
		path = "bash"
	}
	return exec.LookPath(path)
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
		s.Debug = s.makeLogFunc(log.Debug)
	}
	if s.Log == nil {
		s.Log = s.makeLogFunc(log.Verbose)
	}
	if s.Loud == nil {
		s.Loud = s.makeLogFunc(log.Info)
	}
	if s.stdout == nil {
		s.stdout = os.Stderr
	}
	if s.stderr == nil {
		s.stderr = os.Stderr
	}
	return nil
}
