// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
	"github.com/sethvargo/go-envconfig"
)

// ConfigIDFunc can be overridden in tests to provide a stable ID.
var ConfigIDFunc = func(c Config) string { return digest.ID(c) }

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

	// Debug enables debug logging.
	Debug bool `env:"DEBUG"`

	// TargetDir can be used to override the primary target dir.
	TargetDir string `env:"TARGET_DIR"`

	Primary      Paths `env:",prefix=PRIMARY_"`
	Verification Paths `env:",prefix=VERIFICATION_"`

	VerificationResult string `env:"VERIFICATION_RESULT"`
}

type Paths struct {
	// BuildRoot is the absolute path where the instructions are run for this build.
	// We read it from the environment only to support testing.
	BuildRoot string `env:"BUILD_ROOT"`
	// TargetDir can be used to overwrite the target output directory.
	TargetDir string `env:"TARGET_DIR"`
	// BuildResult is the absolute path where the build result will be written.
	// This is the same as the cache path for that build result.
	// We do not read BuildResult from the environment.
	BuildResult string
}

// FromEnvironment creates a new Config from environment variables
// and repository context in the current working directory.
func FromEnvironment(creator crt.Tool, dir string) (Config, error) {
	var c Config
	ctx := context.Background()
	if err := envconfig.Process(ctx, &c); err != nil {
		return c, err
	}

	rc, err := crt.GetRepoContext(dir, build.Dirs.List())
	if err != nil {
		return c, err
	}

	return c.init(rc, creator)
}

// buildConfig returns a BuildConfig based on this Config, rooted at root.
// The root must be an absolute path.
func (c Config) buildConfig(root string) (build.Config, error) {
	// If c.TargetDir is empty here then this option is a no-op,
	// we just pass it through rather than wrapping in a conditional.
	opts := build.WithTargetDir(c.TargetDir)
	paths, err := build.NewBuildPaths(root, c.Product.ExecutableName, c.Parameters.ZipName, opts)
	if err != nil {
		return build.Config{}, err
	}
	var reproducible bool
	if c.Reproducible == "assert" {
		reproducible = true
	}
	return build.NewConfig(c.Product, c.Parameters, paths, c.Tool, reproducible)
}

// PrimaryBuildConfig returns the config for the primary build.
func (c Config) PrimaryBuildConfig() (build.Config, error) {
	return c.buildConfig(c.Primary.BuildRoot)
}

// VerificationBuildConfig returns the config for a verification build.
func (c Config) VerificationBuildConfig() (build.Config, error) {
	return c.buildConfig(c.Verification.BuildRoot)
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

	if strings.ToLower(c.Parameters.OS) == "windows" {
		c.Product.ExecutableName = ensureExtension(c.Product.ExecutableName, ".exe")
	}

	primaryPaths := build.NewPrimaryDirs(c.Product, c.Parameters, creator)

	if c.Primary.BuildResult == "" {
		c.Primary.BuildResult = primaryPaths.BuildResultCacheDir()
	}

	verificationPaths := build.NewVerificationDirs(c.Product, c.Parameters, creator)

	// Default the primary build root to the current directory.
	if c.Primary.BuildRoot == "" {
		c.Primary.BuildRoot = rc.Dir
	}
	if c.Verification.BuildRoot == "" {
		c.Verification.BuildRoot = verificationPaths.RemoteBuildRoot()
	}

	if c.Verification.BuildResult == "" {
		c.Verification.BuildResult = verificationPaths.BuildResultCacheDir()
	}

	c.Tool = creator

	if c.VerificationResult == "" {
		c.VerificationResult = verificationPaths.VerificationResultCachePath(ConfigIDFunc(c), c.Parameters.ZipName)
	}

	return c, nil
}

func ensureExtension(s, ext string) string {
	if strings.HasSuffix(s, ext) {
		return s
	}
	return s + ext
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
