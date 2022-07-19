package crt

// BuildConfig contains the complete configuration to build a single binary
// on a specific host.
type BuildConfig struct {
	// Product is the logical product being built.
	Product Product
	// BuildParameters are the invariant build parameters that must be used
	// in order to reproduce the build.
	Parameters BuildParameters
	// Paths are local to a build on a specific machine.
	Paths BuildPaths
}

// NewBuildConfig expects product, params, and paths to be fully initialized.
func NewBuildConfig(product Product, params BuildParameters, paths BuildPaths) (BuildConfig, error) {
	return BuildConfig{
		Product:    product,
		Parameters: params,
		Paths:      paths,
	}, nil
}
