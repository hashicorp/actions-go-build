package build

import (
	"fmt"
	"strings"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
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
	// Reproducible tells us whether the build is expected to be reproducible.
	// This is used by downstream processes.
	Reproducible bool
}

func ensureExtension(s, ext string) string {
	if strings.HasSuffix(s, ext) {
		return s
	}
	return s + ext
}

// NewConfig expects product, params, and paths to be fully initialized.
func NewConfig(product crt.Product, params Parameters, paths Paths, creator crt.Tool, reproducible bool) (Config, error) {
	c := Config{
		Product:      product,
		Parameters:   params,
		Paths:        paths,
		Tool:         creator,
		Reproducible: reproducible,
	}

	// For windows, append .exe to the bin name if not already there.
	if strings.ToLower(c.Parameters.OS) == "windows" {
		c.Product.ExecutableName = ensureExtension(c.Product.ExecutableName, ".exe")
	}

	return c, nil
}

func CompoundID(things ...any) string {
	ids := make([]string, len(things))
	for i, t := range things {
		ids[i] = ID(t)
	}
	s, err := digest.SHA256HexStrings(ids...)
	if err != nil {
		panic(err)
	}
	return s
}

func ID[T any](thing T) string {
	zero := *(new(T))
	zeroHash, err := digest.JSONSHA256Hex(zero)
	if err != nil {
		panic(err)
	}
	thingHash, err := digest.JSONSHA256Hex(thing)
	if err != nil {
		panic(err)
	}
	if thingHash == zeroHash {
		panic(fmt.Errorf("Can't take ID of zero %T: % #v", thing, thing))
	}
	return thingHash
}

// ID is a unique sha256 hash of this Config.
func (c Config) ID() string {
	return ID(c)
}

func (c Config) BuildResultCachePath(verification bool) string {
	return newDirsFromConfig(c, verification).BuildResultCacheDir()
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
	return newDirsFromConfig(c, true).RemoteBuildRoot()
}

func (c Config) RemotePrimaryRoot() string {
	return newDirsFromConfig(c, false).RemoteBuildRoot()
}
