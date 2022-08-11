package build

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

type td struct {
	kind string
}

type tempDirs struct {
	Primary, Verification td
}

var tempDir = tempDirs{
	Primary:      td{"primary"},
	Verification: td{"verification"},
}

func getTempDirs(verification bool) td {
	if verification {
		return tempDir.Verification
	}
	return tempDir.Primary
}

func assertSourceHash(c Config) {
	if c.Product.SourceHash != "" {
		return
	}
	// It's the maintainers' jobs to make sure we don't hit this panic.
	// It's here to avoid writing undiscoverable files to the cache.
	if (c == Config{}) {
		log.Panicf("SourceHash is empty; Config is empty.")
	}
	log.Panicf("SourceHash is empty; Config is nonempty: % #v", c)
}

func (td td) RemoteBuildRoot(c Config) string {
	assertSourceHash(c)
	return td.cachePath(c, "source", c.Product.SourceHash)
}

func (td td) SourceDownloadPath(c Config) string {
	assertSourceHash(c)
	return td.cachePath(c, "sourcearchive", c.Product.SourceHash)
}

func (td td) BuildResultCachePath(c Config) string {
	assertSourceHash(c)
	return td.cachePath(c, "buildresult", c.Product.SourceHash)
}

func (td td) cachePath(c Config, kind, id string) string {
	return td.tempDirPath(c.Tool, "cache", kind, id)
}

func (td td) tempDirPath(tool crt.Tool, elem ...string) string {
	return prefixPath(elem, os.TempDir(), tool.Name, tool.Version, tool.Revision, td.kind)
}

func prefix(slice []string, with ...string) []string { return append(with, slice...) }

func prefixPath(slice []string, with ...string) string {
	return filepath.Join(prefix(slice, with...)...)
}
