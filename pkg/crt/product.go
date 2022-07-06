package crt

import (
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
	// Name is the product name. This is used to derive the default names
	// for the executable binary, the zip package, deb and rpm packages,
	// and container image tags.
	//
	// For single-product repositories, Name is typically the repository
	// name (i.e. the last path segment of Repository).
	Name string `env:"PRODUCT_NAME"`
	// Version is the base version + prerelease, not including any metadata.
	// It is used alongside Name to derive default names for the zip package,
	// deb and rpm packages, and container image tags.
	Version string `env:"PRODUCT_VERSION"`
	// Revision is the commit SHA of the product being built.
	Revision string
	// RevisionTime is the commit timestamp of Revision in RFC3339 format.
	// This is useful when we need to include a timestamp in a binary whilst
	// keeping the binary reproducible. (It can be used as a sort of "build
	// time").
	RevisionTime string
}

func (p Product) Init(rc RepoContext) Product {
	return p.trimSpace().setDefaults(rc)
}

// trimSpace trims space from the user-provided input fields only.
func (p Product) trimSpace() Product {
	p.Repository = strings.TrimSpace(p.Repository)
	p.Name = strings.TrimSpace(p.Name)
	p.Version = strings.TrimSpace(p.Version)
	return p
}

func (p Product) setDefaults(rc RepoContext) Product {
	if p.Name == "" {
		p.Name = filepath.Base(p.Repository)
	}
	p.Revision = rc.CommitSHA
	p.RevisionTime = rc.CommitTime.UTC().Format(time.RFC3339)
	return p
}
