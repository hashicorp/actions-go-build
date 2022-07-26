package build

import "github.com/hashicorp/actions-go-build/pkg/crt"

// BuildConfig contains the complete configuration to build a single binary
// on a specific host.
type BuildConfig struct {
	// Product is the logical product being built.
	Product crt.Product
	// BuildParameters are the invariant build parameters that must be used
	// in order to reproduce the build.
	Parameters Parameters
	// Paths are local to a build on a specific machine.
	Paths BuildPaths
}

// NewBuildConfig expects product, params, and paths to be fully initialized.
func NewBuildConfig(product crt.Product, params Parameters, paths BuildPaths) (BuildConfig, error) {
	return BuildConfig{
		Product:    product,
		Parameters: params,
		Paths:      paths,
	}, nil
}
