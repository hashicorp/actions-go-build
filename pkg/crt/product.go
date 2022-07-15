package crt

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
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
	// CoreName is the product's core name. This is the same as Name,
	// minus any "-enterprise" suffix. This is a derived value not read from
	// the env directly.
	CoreName string
	// CoreVersion is the base version + prerelease, not including any metadata.
	// It is used alongside Name to derive default names for the zip package,
	// deb and rpm packages, and container image tags.
	CoreVersion string
	// VersionMeta is the metadata portion of the version string.
	VersionMeta string `env:"PRODUCT_VERSION_META"`
	// Version is the full version string made up of CoreVersion + VersionMeta.
	// If this is set externally via the PRODUCT_VERSION variable, then CoreVersion
	// and VersionMeta are disregarded.
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

func (p Product) RevisionTimestamp() (time.Time, error) {
	ts, err := time.Parse(time.RFC3339, p.RevisionTime)
	return ts, maybeErr(err, "invalid revision timestamp %q", p.RevisionTime)
}

// trimSpace trims space from the user-provided input fields only.
func (p Product) trimSpace() Product {
	p.Repository = strings.TrimSpace(p.Repository)
	p.Name = strings.TrimSpace(p.Name)
	p.Version = strings.TrimSpace(p.Version)
	return p
}

func (p Product) setDefaults(rc RepoContext) Product {
	if p.Repository == "" {
		p.Repository = rc.RepoName
	}
	if p.Name == "" {
		p.Name = filepath.Base(p.Repository)
	}
	p.CoreName = strings.TrimSuffix(p.Name, "-enterprise")

	// Figure out the version.
	if p.Version != "" && p.VersionMeta != "" {
		// TODO: Handle this case gracefully by returning an error.
		log.Panicf("both version and version_meta are set")
	}
	if p.Version == "" {
		vfmt := rc.CoreVersion.String()
		if p.VersionMeta != "" {
			vfmt += fmt.Sprintf("+%s", p.VersionMeta)
		}
		// TODO: Handle error instead of using must, and return the error.
		v := version.Must(version.NewVersion(vfmt))
		p.Version = v.String()
	}
	v := version.Must(version.NewVersion(p.Version))
	p.CoreVersion = v.Core().String()
	p.VersionMeta = v.Metadata()

	// Revision things.
	p.Revision = rc.CommitSHA
	p.RevisionTime = rc.CommitTime.UTC().Format(time.RFC3339)
	return p
}

func maybeErr(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
