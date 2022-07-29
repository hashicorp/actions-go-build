package build

import (
	"errors"
	"testing"
)

func result(id string) *Result { return &Result{ErrorMessage: id} }
func assertResultsEqual(t *testing.T, got, want Result) {
	t.Helper()
	if got.ErrorMessage == want.ErrorMessage {
		return
	}
	t.Errorf("got result %q; want %q", got.ErrorMessage, want.ErrorMessage)
}
func makeOpts(opts ...ManagerOption) []ManagerOption { return opts }

var e = errors.New

func TestResultManager_ok(t *testing.T) {

	cases := []struct {
		desc  string
		build *mockBuild
		opts  []ManagerOption
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
		{
			desc:  "cached no prebuild err",
			build: &mockBuild{cached: result("cached")},
			opts:  makeOpts(WithPreBuild(func(Build) error { return e("prebuild") })),
			want:  result("cached"),
		},
	}

	for _, c := range cases {
		build, opts, want := c.build, c.opts, c.want
		t.Run(c.desc, func(t *testing.T) {
			m := NewManager(build, opts...)
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
		opts    []ManagerOption
		wantErr error
	}{
		{
			desc:    "default return cache err",
			build:   &mockBuild{cached: result("cached"), cacheErr: e("cache err")},
			opts:    nil,
			wantErr: e("cache err"),
		},
		{
			desc:    "no cache return prebuild err",
			build:   &mockBuild{},
			opts:    makeOpts(WithPreBuild(func(Build) error { return e("prebuild") })),
			wantErr: e("prebuild"),
		},
		{
			desc:    "no cache return prebuild err",
			build:   &mockBuild{},
			opts:    makeOpts(WithPreBuild(func(Build) error { return e("prebuild") })),
			wantErr: e("prebuild"),
		},
	}

	for _, c := range cases {
		build, opts, wantErr := c.build, c.opts, c.wantErr
		t.Run(c.desc, func(t *testing.T) {
			m := NewManager(build, opts...)
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

func (m *mockBuild) Run() Result {
	if m.fresh != nil {
		return *m.fresh
	}
	return Result{}
}
func (m *mockBuild) Env() []string  { return nil }
func (m *mockBuild) Config() Config { return Config{} }
func (m *mockBuild) CachedResult() (Result, bool, error) {
	if m.cached == nil {
		return Result{}, false, m.cacheErr
	}
	return *m.cached, true, m.cacheErr
}
