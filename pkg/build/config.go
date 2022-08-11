package build

import (
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

// Config contains the complete configuration to build a single binary
// on a specific host.
type Config struct {
	// Product is the logical product being built.
	Product crt.Product
	// BuildParameters are the invariant build parameters that must be used
	// in order to reproduce the build.
	Parameters Parameters
	// Paths are local to a build on a specific machine.
	Paths Paths
	// Tool is info about the tool that created this build.Config.
	Tool crt.Tool
}

// NewConfig expects product, params, and paths to be fully initialized.
func NewConfig(product crt.Product, params Parameters, paths Paths, creator crt.Tool) (Config, error) {
	return Config{
		Product:    product,
		Parameters: params,
		Paths:      paths,
		Tool:       creator,
	}, nil
}

func (c Config) buildResultCachePath(verification bool) string {
	if verification {
		return tempDir.Verification.BuildResultCachePath(c)
	}
	return tempDir.Primary.BuildResultCachePath(c)
}

// ChangeRoot returns a copy of this Config with an updated build root.
func (c Config) ChangeRoot(dir string) (Config, error) {
	var err error
	c.Paths, err = NewBuildPaths(dir, c.Product.ExecutableName, c.Parameters.ZipName)
	return c, err
}

func (c Config) ChangeToVerificationRoot() (Config, error) {
	return c.ChangeRoot(c.VerificationRoot())
}

func (c Config) ChangeToRemotePrimaryRoot() (Config, error) {
	return c.ChangeRoot(c.RemotePrimaryRoot())
}

func (c Config) VerificationRoot() string {
	return tempDir.Verification.RemoteBuildRoot(c)
}

func (c Config) RemotePrimaryRoot() string {
	return tempDir.Primary.RemoteBuildRoot(c)
}
