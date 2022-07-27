package build

import (
	"fmt"
	"log"
	"path/filepath"

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
}

// NewConfig expects product, params, and paths to be fully initialized.
func NewConfig(product crt.Product, params Parameters, paths Paths) (Config, error) {
	return Config{
		Product:    product,
		Parameters: params,
		Paths:      paths,
	}, nil
}

func (c Config) buildResultCachePath() string {
	if c.Product.SourceHash == "" {
		log.Panicf("SourceHash is empty; Config looks like this: % #v", c)
	}
	filename := fmt.Sprintf("buildresult-%s.json", c.Product.SourceHash)
	return filepath.Join(c.Paths.MetaDir, filename)
}
