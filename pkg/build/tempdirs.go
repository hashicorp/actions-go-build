package build

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type TempDirs struct {
	kind    string
	product crt.Product
	tool    crt.Tool
}

func newDirsFromConfig(c Config, verification bool) TempDirs {
	if verification {
		return NewVerificationDirs(c.Product, c.Tool)
	}
	return NewPrimaryDirs(c.Product, c.Tool)
}

func NewPrimaryDirs(p crt.Product, t crt.Tool) TempDirs {
	return NewTempDirs("primary", p, t)
}

func NewVerificationDirs(p crt.Product, t crt.Tool) TempDirs {
	return NewTempDirs("verification", p, t)
}

type tempDirs struct {
	Primary, Verification TempDirs
}

func NewTempDirs(kind string, p crt.Product, t crt.Tool) TempDirs {
	assertSourceHash(p)
	return TempDirs{kind: kind, product: p, tool: t}
}

func TempDirsFromConfig(c Config) tempDirs {
	return tempDirs{
		Primary:      NewPrimaryDirs(c.Product, c.Tool),
		Verification: NewVerificationDirs(c.Product, c.Tool),
	}
}

func assertSourceHash(p crt.Product) {
	if p.SourceHash != "" {
		return
	}
	// It's the maintainers' jobs to make sure we don't hit this panic.
	// It's here to avoid writing undiscoverable files to the cache.
	if (p == crt.Product{}) {
		log.Panicf("SourceHash is empty; Product is empty.")
	}
	log.Panicf("SourceHash is empty; Product is nonempty: % #v", p)
}

func (d TempDirs) RemoteBuildRoot() string {
	return d.cacheDir("source")
}

func (d TempDirs) SourceDownloadDir() string {
	return d.cacheDir("sourcearchive")
}

func (d TempDirs) BuildResultCacheDir() string {
	return d.cacheDir("buildresult")
}

func (d TempDirs) cacheDir(kind string) string {
	return d.tempDirPath(d.tool, "cache", kind, d.product.SourceHash)
}

func (d TempDirs) tempDirPath(tool crt.Tool, elem ...string) string {
	return prefixPath(elem, os.TempDir(), tool.Name, tool.Version, tool.Revision, d.kind)
}

func prefix(slice []string, with ...string) []string { return append(with, slice...) }

func prefixPath(slice []string, with ...string) string {
	return filepath.Join(prefix(slice, with...)...)
}
