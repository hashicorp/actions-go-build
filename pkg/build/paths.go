package build

import (
	"fmt"
	"path/filepath"
)

type DirNames struct {
	Target, Zip, Meta string
}

var Dirs = DirNames{"dist", "out", "meta"}

func (ds DirNames) List() []string {
	return []string{ds.Target, ds.Zip, ds.Meta}
}

// Paths are host-specific absolute paths to various things. We need to
// be aware of these paths in order to be able to do comparisons between
// primary and verification builds. They must not affect the bytes produced.
type Paths struct {
	// WorkDir is the absolute directory to run the build instructions in.
	WorkDir string
	// TargetDir is the absolute path to the dir where any other files
	// needed to be included in the zip file should be placed.
	TargetDir string
	// BinPath is the absolute path to the executable binary the instructions
	// must create.
	BinPath string
	// ZipPath is the path to the zip file that will be created.
	ZipPath string
	// MetaDir is where we write metadata about this build.
	MetaDir string
}

func NewBuildPaths(root, executableName, zipName string) (Paths, error) {
	var bp Paths
	if !filepath.IsAbs(root) {
		return bp, fmt.Errorf("root path %q is not absolute", root)
	}
	return bp.setDefaults(root, executableName, zipName), nil
}

func (bp Paths) ZipDir() string {
	return filepath.Dir(bp.ZipPath)
}

func (bp Paths) trimSpace() Paths {
	// Placeholder func for consistency.
	return bp
}

func (bp Paths) setDefaults(root, executableName, zipName string) Paths {
	bp.WorkDir = root
	bp.TargetDir = filepath.Join(root, Dirs.Target)
	bp.BinPath = filepath.Join(root, Dirs.Target, executableName)
	bp.ZipPath = filepath.Join(root, Dirs.Zip, zipName)
	bp.MetaDir = filepath.Join(root, Dirs.Meta)
	return bp
}
