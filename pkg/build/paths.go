// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	// BinPath is the absolute path to the executable binary the instructions
	// must create.
	BinPath string
	// ZipPath is the path to the zip file that will be created.
	ZipPath string
	// MetaDir is where we write metadata about this build.
	MetaDir string
}

type pathsSettings struct {
	// targetDir is used to calculate the default bin path.
	targetDir string
}

// buildPathsSettings contains optional settings for build paths.
type buildPathsSettings struct {
	targetDir string
}

type BuildPathsOpt func(s *pathsSettings)

func WithTargetDir(path string) BuildPathsOpt {
	return func(s *pathsSettings) { s.targetDir = path }
}

func NewBuildPaths(root, executableName, zipName string, opts ...BuildPathsOpt) (Paths, error) {
	var bp Paths
	if !filepath.IsAbs(root) {
		return bp, fmt.Errorf("root path %q is not absolute", root)
	}

	settings := pathsSettings{}
	// Apply options.
	for _, o := range opts {
		o(&settings)
	}

	return bp.setDefaults(root, executableName, zipName, settings), nil
}

func (bp Paths) ZipDir() string {
	return filepath.Dir(bp.ZipPath)
}

// TargetDir is the absolute path to the dir where any other files
// needed to be included in the zip file should be placed.
func (bp Paths) TargetDir() string {
	return filepath.Dir(bp.BinPath)
}

func (bp Paths) trimSpace() Paths {
	// Placeholder func for consistency.
	return bp
}

func (bp Paths) setDefaults(root, executableName, zipName string, s pathsSettings) Paths {
	if len(bp.WorkDir) == 0 {
		bp.WorkDir = root
	}
	// Set the internal targetDir if it wasn't already set by an option.
	if len(s.targetDir) == 0 {
		s.targetDir = filepath.Join(bp.WorkDir, Dirs.Target)
	}
	if len(bp.BinPath) == 0 {
		bp.BinPath = filepath.Join(s.targetDir, executableName)
	}
	if len(bp.ZipPath) == 0 {
		bp.ZipPath = filepath.Join(bp.WorkDir, Dirs.Zip, zipName)
	}
	if len(bp.MetaDir) == 0 {
		bp.MetaDir = filepath.Join(bp.WorkDir, Dirs.Meta)
	}
	return bp
}
