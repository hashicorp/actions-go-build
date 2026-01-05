// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

func result(id string) *Result { return &Result{ErrorMessage: id} }

func assertResultsEqual(t *testing.T, got, want Result) {
	t.Helper()
	if strings.HasPrefix(got.ErrorMessage, want.ErrorMessage) {
		return
	}
	t.Errorf("got result %q; want %q", got.ErrorMessage, want.ErrorMessage)
}

func makeOpts(opts ...Option) []Option { return opts }

var e = errors.New

// TODO: This test needs refactoring now we have a separate Runner type.
// Manager -> Runner -> Build is the hierarchy, but this is testing the
// manager in terms of builds, so skipping the runner level.
//
// There should be separate tests for Runner -> Build and Manager -> Runner.
//
// The current shape of these tests is caused by refactoring the main codebase
// but not the tests at the same time.
func TestManager_ok(t *testing.T) {

	cases := []struct {
		desc  string
		build *mockBuild
		opts  []Option
		want  *Result
	}{
		{
			desc:  "default return cached",
			build: &mockBuild{fresh: result("fresh"), cached: result("cached")},
			opts:  nil,
			want:  result("cached"),
		},
		{
			desc:  "default return fresh if no cache",
			build: &mockBuild{fresh: result("fresh"), cached: nil},
			opts:  nil,
			want:  result("fresh"),
		},
		{
			desc:  "force-run return fresh ignore cache",
			build: &mockBuild{fresh: result("fresh"), cached: result("cached")},
			opts:  makeOpts(WithForceRebuild(true)),
			want:  result("fresh"),
		},
		{
			desc:  "force-run return fresh",
			build: &mockBuild{fresh: result("fresh"), cached: nil},
			opts:  makeOpts(WithForceRebuild(true)),
			want:  result("fresh"),
		},
		{
			desc:  "force-build no cache err",
			build: &mockBuild{fresh: result("fresh"), cacheErr: e("cache err")},
			opts:  makeOpts(WithForceRebuild(true)),
			want:  result("fresh"),
		},
	}

	for _, c := range cases {
		build, opts, want := c.build, c.opts, c.want
		t.Run(c.desc, func(t *testing.T) {
			runner, err := NewRunner(build)
			if err != nil {
				t.Fatal(err)
			}

			m, err := NewManager(runner, opts...)
			if err != nil {
				t.Fatal(err)
			}
			got, err := m.Result()
			if err != nil {
				t.Fatal(err)
			}
			assertResultsEqual(t, got, *want)
		})
	}
}

func TestManager_err(t *testing.T) {
	cases := []struct {
		desc    string
		build   *mockBuild
		opts    []Option
		wantErr error
	}{
		{
			desc:    "default return cache err",
			build:   &mockBuild{cached: result("cached"), cacheErr: e("cache err1")},
			opts:    nil,
			wantErr: e("inspecting cache: cache err1"),
		},
	}

	for _, c := range cases {
		build, opts, wantErr := c.build, c.opts, c.wantErr
		t.Run(c.desc, func(t *testing.T) {
			runner, err := NewRunner(build)
			if err != nil {
				t.Fatal(err)
			}
			m, err := NewManager(runner, opts...)
			if err != nil {
				t.Fatal(err)
			}
			_, gotErr := m.Result()
			if gotErr == nil {
				t.Fatalf("got nil error; want %q", wantErr)
			}
			want := wantErr.Error()
			got := gotErr.Error()
			if got != want {
				t.Errorf("got error %q; want %q", got, want)
			}
		})
	}
}

type mockBuild struct {
	fresh, cached *Result
	cacheErr      error
}

func (m *mockBuild) Env() []string                   { return nil }
func (m *mockBuild) ChangeRoot(string) error         { return nil }
func (m *mockBuild) ChangeToVerificationRoot() error { return nil }
func (m *mockBuild) Kind() string                    { return "mock" }
func (m *mockBuild) IsVerification() bool            { return false }
func (m *mockBuild) Dirs() TempDirs {
	return NewTempDirs("test", crt.Product{SourceHash: "deadbeef"}, Parameters{}, crt.Tool{})
}

func (m *mockBuild) Config() Config {
	return Config{
		Product: crt.Product{
			SourceHash: "blargle",
		},
		Parameters: Parameters{
			Instructions: "test",
		},
		Paths: Paths{
			MetaDir: TempDirFunc(),
		},
	}
}
func (m *mockBuild) CachedResult() (Result, bool, error) {
	if m.cached == nil {
		return Result{}, false, m.cacheErr
	}
	return *m.cached, true, m.cacheErr
}
func (m *mockBuild) Steps() []Step {
	return []Step{
		newStep("fresh", func() error { return errors.New("an error") }),
	}
}
