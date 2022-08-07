package build

// Manager is responsible for orchestrating the running of builds.
// By default it will return cached build results rather then re-running
// a build that's already been done.
type Manager struct {
	Settings
	runner *Runner
}

func NewManager(r *Runner, opts ...Option) (*Manager, error) {
	s, err := newSettings("manager", opts)
	if err != nil {
		return nil, err
	}
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
	bm.Debug("Inspecting cache for build result.")
	return bm.runner.build.CachedResult()
}

// Result returns the build result either from cache if present and not forcing a rebuild,
// or by running the build to generate a new result. The only errors that can be returned are
// from the attempt to load from cache, so to check if the build failed or not you still need
// to call the Result's Error method.
func (bm *Manager) Result() (Result, error) {
	if bm.forceRebuild {
		bm.Debug("Force-rebuild on, not inspecting cache.")
	} else {
		r, cached, err := bm.ResultFromCache()
		if err != nil {
			return r, err
		}
		if cached {
			bm.Debug("Loaded build result from cache.")
			return r, nil
		}
		bm.Debug("No build result avilable in cache.")
	}
	bm.Debug("Running a fresh build...")
	result := bm.runner.Run()
	cachePath, err := result.Save()
	if err != nil {
		return result, err
	}
	bm.Debug("Saved result to cache: %q", cachePath)
	return result, nil
}
