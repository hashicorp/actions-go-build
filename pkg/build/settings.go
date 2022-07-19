package build

import (
	"context"
	"io"
	"os"
)

// Settings contains settings for running the instructions.
// These are not to be confused with crt.BuildConfig, these settings
// are build-run specific and not part of the _definition_ of the build.
// Don't use this directly, use the With... functions to set
// settings when calling New.
type Settings struct {
	bash    string
	context context.Context
	stdout  io.Writer
	stderr  io.Writer
}

func newSettings(options []Option) (Settings, error) {
	out := &Settings{}
	for _, o := range options {
		o(out)
	}
	if err := out.setDefaults(); err != nil {
		return Settings{}, err
	}
	return *out, nil
}

func (s *Settings) setDefaults() (err error) {
	s.bash, err = resolveBashPath(s.bash)
	if err != nil {
		return err
	}
	if s.context == nil {
		s.context = context.Background()
	}
	if s.stdout == nil {
		s.stdout = os.Stdout
	}
	if s.stderr == nil {
		s.stderr = os.Stderr
	}
	return nil
}

// Option represents a function that configures Settings.
type Option func(*Settings)

func WithContext(c context.Context) Option { return func(s *Settings) { s.context = c } }
func WithStdout(w io.Writer) Option        { return func(s *Settings) { s.stdout = w } }
func WithStderr(w io.Writer) Option        { return func(s *Settings) { s.stderr = w } }
func WithBash(path string) Option          { return func(s *Settings) { s.bash = path } }
