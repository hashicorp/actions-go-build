package build

import "log"

type logFunc func(string, ...any)

// Manager is responsible for orchestrating the running of builds.
// By default it will return cached build results rather then re-running
// a build that's already been done.
type Manager struct {
	runner       *Runner
	forceRebuild bool
	log, debug   logFunc
}

type ManagerOption func(*Manager)

func NewManager(r *Runner, opts ...ManagerOption) *Manager {
	noopLogFunc := func(string, ...any) {}
	bm := &Manager{runner: r, log: log.Printf, debug: noopLogFunc}
	for _, o := range opts {
		o(bm)
	}
	return bm
}

func WithForceRebuild(on bool) ManagerOption           { return func(bm *Manager) { bm.forceRebuild = on } }
func WithLogFunc(f func(string, ...any)) ManagerOption { return func(bm *Manager) { bm.log = f } }
func WithDebugLogFunc(f func(string, ...any)) ManagerOption {
	return func(bm *Manager) { bm.debug = f }
}

func (bm *Manager) ResultFromCache() (Result, bool, error) {
	bm.debug("Inspecting cache for build result.")
	return bm.runner.build.CachedResult()
}

// Result returns the build result either from cache if present and not forcing a rebuild,
// or by running the build to generate a new result. The only errors that can be returned are
// from the attempt to load from cache, so to check if the build failed or not you still need
// to call the Result's Error method.
func (bm *Manager) Result() (Result, error) {
	if bm.forceRebuild {
		bm.debug("Force-rebuild on, not inspecting cache.")
	} else {
		if r, cached, err := bm.ResultFromCache(); cached || err != nil {
			if cached {
				bm.log("Loaded build result from cache.")
			}
			return r, err
		}
	}
	bm.log("Running a fresh build.")
	result := bm.runner.Run()
	cachePath, err := result.Save()
	if err != nil {
		return result, err
	}
	bm.debug("Saved result to cache: %q", cachePath)
	return result, nil
}
