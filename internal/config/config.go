package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/sethvargo/go-envconfig"
)

// Config represents the action configuration.
type Config struct {
	// Product contains the invariant product details, which don't
	// vary between different builds of the same commit. (I.e. this
	// doesn't contain platform- or build-specific info, like OS/ARCH,
	// or build tags, etc.)
	Product crt.Product

	// BuildParameters are invariant build details, required alongside
	// the Product definition to capture the full instructions needed to
	// reproduce a build.
	Parameters build.Parameters

	// Reproducible tells the action whether this build ought to be reproducible.
	// It must be one of these three values:
	//   - "assert" - this build must be reproducible, fail otherwise.
	//   - "report" - run the verification build and report, but don't fail the build.
	//   - "nope"   - don't run the verification build at all.
	Reproducible string `env:"REPRODUCIBLE"`

	// Tool is the version of actions-go-build that created this config.
	Tool crt.Tool

	// Optional inputs which do not affect the bytes produced.
	// Mostly useful for testing.

	// PrimaryBuildRoot is the absolute path where the instructions are run for the
	// primary build. This path should already exist and contain the product repo
	// checked out at the commit we want to build.
	// Default: current working directory.
	PrimaryBuildRoot string `env:"PRIMARY_BUILD_ROOT"`
	// VerificationBuildRoot is the absolute path where the instructions are run
	// for the verification build. This path should not already exist, it is created
	// by making a recursive copy of the primary build root.
	// Default: a newly minted temporary directory.
	VerificationBuildRoot string `env:"VERIFICATION_BUILD_ROOT"`
}

// FromEnvironment creates a new Config from environment variables
// and repository context in the current working directory.
func FromEnvironment(creator crt.Tool) (Config, error) {
	var c Config
	ctx := context.Background()
	if err := envconfig.Process(ctx, &c); err != nil {
		return c, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return c, err
	}

	rc, err := crt.GetRepoContext(wd, build.Dirs.List())
	if err != nil {
		return c, err
	}

	return c.init(rc, creator)
}

// buildConfig returns a BuildConfig based on this Config, rooted at root.
// The root must be an absolute path.
func (c Config) buildConfig(root string) (build.Config, error) {
	paths, err := build.NewBuildPaths(root, c.Product.ExecutableName, c.Parameters.ZipName)
	if err != nil {
		return build.Config{}, err
	}
	return build.NewConfig(c.Product, c.Parameters, paths, c.Tool)
}

// PrimaryBuildConfig returns the config for the primary build.
func (c Config) PrimaryBuildConfig() (build.Config, error) {
	return c.buildConfig(c.PrimaryBuildRoot)
}

// VerificationBuildConfig returns the config for the verification build.
func (c Config) VerificationBuildConfig() (build.Config, error) {
	return c.buildConfig(c.VerificationBuildRoot)
}

func (c Config) init(rc crt.RepoContext, creator crt.Tool) (Config, error) {
	var err error
	if c.Product, err = c.Product.Init(rc); err != nil {
		return c, err
	}
	if c.Parameters, err = c.Parameters.Init(c.Product); err != nil {
		return c, err
	}
	if c.Reproducible, err = c.resolveReproducible(); err != nil {
		return c, err
	}
	if c.PrimaryBuildRoot == "" {
		c.PrimaryBuildRoot = rc.Dir
	}
	if c.VerificationBuildRoot == "" {
		c.VerificationBuildRoot = defaultVerificationBuildRoot(rc)
	}
	c.Tool = creator
	return c, nil
}

func defaultVerificationBuildRoot(rc crt.RepoContext) string {
	return filepath.Join(os.TempDir(), "actions-go-build", rc.RepoName, rc.SourceHash, "verification")
}

func (c Config) resolveReproducible() (string, error) {
	switch c.Reproducible {
	default:
		return "", fmt.Errorf("%q is not a valid value for 'reproducible', must be one of 'assert' (default), 'report', or 'nope')", c.Reproducible)
	case "":
		return "assert", nil
	case "assert", "report", "nope":
		return c.Reproducible, nil
	}
}
