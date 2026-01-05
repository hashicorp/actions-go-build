// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
)

type TempDirs struct {
	cacheKey
	kind string
}

type cacheKey struct {
	product    crt.Product
	parameters Parameters
	tool       crt.Tool
}

// TempDirFunc is the function used by this package to get the system temp dir.
// You can override this for testing purposes to get platform-independent paths.
var TempDirFunc = os.TempDir

// CacheKeyFunc can be overridden by tests to generate stable strings.
var CacheKeyFunc = digest.CompoundID

func (ck cacheKey) Key() string { return CacheKeyFunc(ck.product, ck.parameters, ck.tool) }

func newDirsFromConfig(c Config, verification bool) TempDirs {
	if verification {
		return NewVerificationDirs(c.Product, c.Parameters, c.Tool)
	}
	return NewPrimaryDirs(c.Product, c.Parameters, c.Tool)
}

func NewPrimaryDirs(p crt.Product, params Parameters, t crt.Tool) TempDirs {
	return NewTempDirs("primary", p, params, t)
}

func NewVerificationDirs(p crt.Product, params Parameters, t crt.Tool) TempDirs {
	return NewTempDirs("verification", p, params, t)
}

type tempDirs struct {
	Primary, Verification TempDirs
}

func NewTempDirs(kind string, p crt.Product, params Parameters, t crt.Tool) TempDirs {
	assertSourceHash(p)
	key := cacheKey{p, params, t}
	return TempDirs{kind: kind, cacheKey: key}
}

func assertSourceHash(p crt.Product) {
	if p.SourceHash != "" {
		return
	}
	// It's the maintainers' jobs to make sure we don't hit this panic.
	// It's here to avoid writing undiscoverable files to the cache.
	log.Panicf("SourceHash is empty in product: % #v", p)
}

func (d TempDirs) RemoteBuildRoot(extension ...string) string {
	return d.cacheDir("source", extension...)
}

func (d TempDirs) SourceDownloadDir() string {
	return d.cacheDir("sourcearchive")
}

func (d TempDirs) BuildResultCacheDir(extension ...string) string {
	return d.cacheDir("buildresult", extension...)
}

func (d TempDirs) VerificationResultCachePath(configID, zipName string) string {
	return d.cacheDir("verificationresult", configID, zipName+".json")
}

func (d TempDirs) cacheDir(kind string, extension ...string) string {
	sh := d.product.SourceHash
	// If the sourcehash is a long hex string, we shorten it and prefix with
	// dirty_ or clean_. This doesn't matter for cache invalidation as the
	// hash of the build inputs already includes this.
	matched, err := regexp.Match("[a-f0-9]{40}.*", []byte(sh))
	if err != nil {
		// This panic will only happen if the code above is wrong (i.e. the
		// regexp doesn't compile). So this won't happen in releases since
		// they will have passed tests that assert this panic can't happen.
		panic(err)
	}
	if matched {
		sh = sh[:8]
		if d.product.IsDirty() {
			sh = fmt.Sprintf("dirty_%s", sh)
		} else {
			sh = fmt.Sprintf("clean_%s", sh)
		}
	}
	return d.tempDirPath(prefix(extension, "cache", kind, d.product.Repository, d.product.Name, sh)...)
}

func (d TempDirs) tempDirPath(elem ...string) string {
	return prefixPath(elem, TempDirFunc(), d.tool.Name, d.tool.Version, d.tool.Revision, d.kind, d.Key())
}

func productIDSegments(p crt.Product) []string {
	return []string{p.Repository, p.Name, p.Version.Full}
}

func prefix(slice []string, with ...string) []string { return append(with, slice...) }

func prefixPath(slice []string, with ...string) string {
	return filepath.Join(prefix(slice, with...)...)
}
