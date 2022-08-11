package build

import (
	"fmt"
)

// Manager is responsible for orchestrating the running of builds.
// By default it will return cached build results rather then re-running
// a build that's already been done.
type Manager struct {
	Settings
	runner *Runner
}

func NewManager(r *Runner, opts ...Option) (*Manager, error) {
	s, err := newSettings(opts)
	if err != nil {
		return nil, err
	}
	s.Debug("Initialised")
	return &Manager{
		Settings: s,
		runner:   r,
	}, nil
}

func (bm *Manager) Runner() *Runner {
	return bm.runner
}

func (bm *Manager) Build() Build {
	return bm.runner.build
}

func (bm *Manager) ResultFromCache() (Result, bool, error) {
	return bm.runner.build.CachedResult()
}

// Result returns the build result either from cache if present and not forcing a rebuild,
// or by running the build to generate a new result. The only errors that can be returned are
// from the attempt to load from cache, so to check if the build failed or not you still need
// to call the Result's Error method.
func (bm *Manager) Result() (Result, error) {
	bm.Debug("getting result")
	if bm.forceRebuild {
		bm.Log("Force-rebuild on, running a fresh build...")
		return bm.runBuild()
	}
	bm.Debug("Inspecting cache.")
	r, cached, err := bm.ResultFromCache()
	if err != nil {
		return r, fmt.Errorf("inspecting cache: %w", err)
	}
	if cached {
		bm.Log("Loaded build result from cache; SourceID: %s; Dirty: %t", r.Config.Product.SourceHash, r.Config.Product.IsDirty())
		return r, nil
	}
	bm.Log("No build result avilable in cache; Running a fresh build...")
	return bm.runBuild()
}

func (bm *Manager) runBuild() (Result, error) {
	result := bm.runner.Run()
	cachePath, err := result.Save(bm.Build().IsVerification())
	if err != nil {
		return result, err
	}
	bm.Debug("Saved result to cache: %q", cachePath)
	return result, nil
}
