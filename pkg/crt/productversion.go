package crt

import (
	"strings"

	"github.com/hashicorp/go-version"
)

type ProductVersion struct {
	// Full is the full version string made up of Core + Meta.
	// If this is set externally, then `Meta` must not also be set, if it is
	// that's a validation error.
	Full string `env:"PRODUCT_VERSION"`
	// Core is the base version + prerelease, not including any metadata.
	// It is used alongside Name to derive default names for the zip package,
	// deb and rpm packages, and container image tags.
	Core string
	// Meta is the metadata portion of the version string.
	Meta string `env:"PRODUCT_VERSION_META"`
}

func (p ProductVersion) String() string { return p.Full }

// NewProductVersion accepts a coreVersion and a fullVersion. The fullVersion may
// be empty, in which case the full version will be set to coreVersion-local indicating
// a "local" build, without proper version information.
// If the coreVersion is also empty, then we set it to 0.0.0-unversioned.
func NewProductVersion(coreVersion, fullVersion string) (ProductVersion, error) {
	fullVersion = strings.TrimSpace(fullVersion)
	coreVersion = strings.TrimSpace(coreVersion)
	return ProductVersion{
		Full: fullVersion,
		Core: coreVersion,
	}.Init()
}

// trimSpace trims leading and trailing space from fields read from external inputs
// only.
func (pv ProductVersion) trimSpace() ProductVersion {
	pv.Full = strings.TrimSpace(pv.Full)
	pv.Meta = strings.TrimSpace(pv.Meta)
	return pv
}

// Init ensures that all fields in the ProductVersion are consistent, and fills in
// any missing fields.
func (pv ProductVersion) Init() (ProductVersion, error) {
	if pv.Full != "" {
		v, err := version.NewVersion(pv.Full)
		if err != nil {
			return pv, err
		}
		pv.Core = v.Core().String()
		if p := v.Prerelease(); p != "" {
			pv.Core += "-" + p
		}
		pv.Meta = v.Metadata()
		return pv, nil
	}

	if pv.Core == "" && pv.Meta == "" {
		pv.Meta = "local"
	}

	if pv.Core == "" {
		pv.Full = "0.0.0-unversioned"
		pv.Core = "0.0.0-unversioned"
	}

	pv.Full = pv.Core

	if pv.Meta != "" {
		pv.Full += "+" + pv.Meta
	}

	return pv, nil
}

func (pv ProductVersion) InitWithCoreVersion(coreVersion string) (ProductVersion, error) {
	pv.Core = coreVersion
	return pv.Init()
}
