// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package build

import (
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

// NewConfig expects product, params, and paths to be fully initialized.
func NewConfig(product crt.Product, params Parameters, paths Paths, creator crt.Tool, reproducible bool) (Config, error) {
	c := Config{
		Product:      product,
		Parameters:   params,
		Paths:        paths,
		Tool:         creator,
		Reproducible: reproducible,
	}

	return c, nil
}

// ConfigIDFunc can be overridden in tests to generate stable config IDs.
var ConfigIDFunc = func(c Config) string {
	return digest.ID(c)
}

// ID is a unique sha256 hash of this Config.
func (c Config) ID() string {
	return ConfigIDFunc(c)
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
