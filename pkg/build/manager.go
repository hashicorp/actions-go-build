package build

import "log"

type PreBuildFunc func(Build) error

type logFunc func(string, ...any)

// Manager is responsible for orchestrating the running of builds.
// By default it will return cached build results rather then re-running
// a build that's already been done.
type Manager struct {
	build        Build
	forceRebuild bool
	preBuild     PreBuildFunc
	log, debug   logFunc
}

type ManagerOption func(*Manager)

func NewManager(b Build, opts ...ManagerOption) *Manager {
	noopLogFunc := func(string, ...any) {}
	bm := &Manager{build: b, log: log.Printf, debug: noopLogFunc}
	for _, o := range opts {
		o(bm)
	}
	return bm
}

func WithForceRebuild(on bool) ManagerOption           { return func(bm *Manager) { bm.forceRebuild = on } }
func WithPreBuild(f PreBuildFunc) ManagerOption        { return func(bm *Manager) { bm.preBuild = f } }
func WithLogFunc(f func(string, ...any)) ManagerOption { return func(bm *Manager) { bm.log = f } }
func WithDebugLogFunc(f func(string, ...any)) ManagerOption {
	return func(bm *Manager) { bm.debug = f }
}

// Result returns the build result either from cache if present and not forcing a rebuild,
// or by running the build to generate a new result. The only errors that can be returned are
// from the attempt to load from cache, or from the pre-build func if one is specified.
// So to check if the build failed or not you still need to call the Build's Error method.
func (bm *Manager) Result() (Result, error) {
	if bm.forceRebuild {
		bm.debug("Force-rebuild on, not inspecting cache.")
	} else {
		bm.debug("Inspecting cache for build result.")
		if r, cached, err := bm.build.CachedResult(); cached || err != nil {
			if cached {
				bm.log("Loaded build result from cache.")
			}
			return r, err
		}
	}
	if bm.preBuild != nil {
		bm.debug("Running pre-build function.")
		if err := bm.preBuild(bm.build); err != nil {
			return Result{}, err
		}
	} else {
		bm.debug("No pre-build function defined.")
	}
	bm.log("Running a fresh build.")
	return bm.build.Run(), nil
}
