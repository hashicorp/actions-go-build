// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package product

import (
	_ "embed"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

var (
	// The below vars should all be injected via -ldflags "-X ./product/VAR_NAME=VALUE" ...

	// Repository is the repository name, e.g. github.com/hashicorp/<product_name>
	Repository string
	// Module is the go module housing this product.
	Module string
	// Name is the full product name, typically the same as the last path segment of
	// the repository.
	Name string
	// CoreName is the same as Name minus any -enterprise suffix.
	CoreName string
	// ExecutableName is the canonical name of this executable.
	ExecutableName string
	// VersionFull is the full version string for this product, including prerelease and
	// metadata fields.
	VersionFull string
	// VersionCore is the same as Version minus any metadata field.
	VersionCore string
	// VersionMeta is the metadata field (if any) from VersionFull.
	VersionMeta string
	// Revision is the Git commit SHA of this product. It should be the full 40-char SHA1.
	Revision string
	// RevisionTime is the timestamp of the Revision.
	RevisionTime string
	// SourceHash is either the Revision if this is a clean build (with no uncommitted
	// changes), or else it's a hash of the HEAD commit SHA along with the full contents
	// of any new or changed files (i.e. a dirty build).
	SourceHash string
)

func Product(defaultName, coreVersion string) (crt.Product, error) {
	// If the name was injected, leave it as-is, otherwise
	// use the name supplied. The build process is the source
	// of truth for what the product is, but for local builds
	// we need defaults.
	if Name == "" {
		Name = defaultName
	}
	if Revision == "" {
		Revision = "unknown-revision"
	}
	if SourceHash == "" {
		SourceHash = "unknown-source-hash"
	}
	version, err := crt.ProductVersion{
		Full: VersionFull,
		Core: VersionCore,
		Meta: VersionMeta,
	}.Init()
	return crt.Product{
		Repository:     Repository,
		Module:         Module,
		Name:           Name,
		CoreName:       CoreName,
		ExecutableName: ExecutableName,
		Version:        version,
		Revision:       Revision,
		RevisionTime:   RevisionTime,
		SourceHash:     SourceHash,
	}, err
}
