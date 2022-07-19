package crt

import (
	"fmt"
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

// trimSpace trims leading and trailing space from fields read from external inputs
// only.
func (pv ProductVersion) trimSpace() ProductVersion {
	pv.Full = strings.TrimSpace(pv.Full)
	pv.Meta = strings.TrimSpace(pv.Meta)
	return pv
}

// resolve takes a ProductVersion formed by external inputs and ensures all fields
// are consistently populated.
func (pv ProductVersion) resolve(rc RepoContext) (ProductVersion, error) {
	if pv.Full != "" && pv.Meta != "" {
		return pv, fmt.Errorf("both version and version_meta are set")
	}
	if pv.Full == "" {
		// rc.CoreVersion is the version from the version file.
		vfmt := rc.CoreVersion.String()
		if pv.Meta != "" {
			vfmt += fmt.Sprintf("+%s", pv.Meta)
		}
		v, err := version.NewVersion(vfmt)
		if err != nil {
			return pv, fmt.Errorf("parsing version %q: %w", vfmt, err)
		}
		pv.Full = v.String()
	}
	v, err := version.NewVersion(pv.Full)
	if err != nil {
		return pv, fmt.Errorf("parsing version %q: %w", pv.Full, err)
	}
	pv.Core = v.Core().String()
	pv.Meta = v.Metadata()
	return pv, nil
}
