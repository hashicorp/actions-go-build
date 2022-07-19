package crt

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Product represents a single logical product. There may be multiple
// products per repository. One product typically maps to multiple
// Go binaries (one per platform). It may also map to different binaries
// based on different build parameters e.g. build tags being used.
type Product struct {
	// Repository is the product repository URL minus the scheme.
	// E.g. github.com/hashicorp/lockbox
	Repository string `env:"PRODUCT_REPOSITORY"`
	// Directory is the path to the directory housing this product, inside Repository,
	// relative to the repository root. Defaults to the current working directory.
	// Note that this is optional, you can have multiple products sharing the same
	// directory if wanted. If set, the Instructions should be run inside this directory.
	Directory string `env:"PRODUCT_DIRECTORY"`
	// Name is the product name. This is used to derive the default names
	// for the executable binary, the zip package, deb and rpm packages,
	// container image tags, and other artifacts in the future.
	//
	// Name defaults to the last path segment of Repository, when running
	// in the root of the repository. Otherwise it defaults to the base name of
	// the working directory in which the product is built.
	Name string `env:"PRODUCT_NAME"`
	// CoreName is the product's core name. This is the same as Name,
	// minus any "-enterprise" suffix. This is a derived value not read from
	// the env directly.
	CoreName string
	// ExecutableName is the name of the executable binary representing this
	// product. Defaults to CoreName.
	ExecutableName string `env:"BIN_NAME"`
	// Version is the version of the product being built.
	Version ProductVersion
	// Revision is the commit SHA of the product being built.
	Revision string
	// RevisionTime is the commit timestamp of Revision in RFC3339 format.
	// This is useful when we need to include a timestamp in a binary whilst
	// keeping the binary reproducible. (It can be used as a sort of "build
	// time").
	RevisionTime string
}

func (p Product) Init(rc RepoContext) (Product, error) {
	return p.trimSpace().setDefaults(rc)
}

func (p Product) RevisionTimestamp() (time.Time, error) {
	ts, err := time.Parse(time.RFC3339, p.RevisionTime)
	return ts, maybeErr(err, "invalid revision timestamp %q", p.RevisionTime)
}

// trimSpace trims space from the user-provided input fields only.
func (p Product) trimSpace() Product {
	p.Repository = strings.TrimSpace(p.Repository)
	p.Directory = strings.TrimSpace(p.Directory)
	p.Name = strings.TrimSpace(p.Name)
	p.Version = p.Version.trimSpace()
	p.ExecutableName = strings.TrimSpace(p.ExecutableName)
	return p
}

func (p Product) setDefaults(rc RepoContext) (Product, error) {
	if p.Repository == "" {
		p.Repository = rc.RepoName
	}

	if p.Name == "" {
		p.Name = p.defaultProductName(rc)
	}

	p.CoreName = strings.TrimSuffix(p.Name, "-enterprise")

	if p.ExecutableName == "" {
		p.ExecutableName = p.CoreName
	}

	var err error
	if p.Version, err = p.Version.resolve(rc); err != nil {
		return p, err
	}

	p.Revision = rc.CommitSHA
	p.RevisionTime = rc.CommitTime.UTC().Format(time.RFC3339)
	return p, nil
}

func (p Product) defaultProductName(rc RepoContext) string {
	// If we're in the repo root, use the repo name.
	if rc.Dir == rc.RootDir {
		return filepath.Base(p.Repository)
	}
	// Otherwise use the subdirectory name.
	return filepath.Base(rc.Dir)
}

func maybeErr(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
