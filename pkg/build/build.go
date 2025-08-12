// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/actions-go-build/internal/zipper"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

// Build represents the build of a single binary.
// It could be a primary build or a verification build, this Build doesn't
// need to know.
type Build interface {
	Env() []string
	Config() Config
	CachedResult() (Result, bool, error)
	Steps() []Step
	ChangeRoot(string) error
	ChangeToVerificationRoot() error
	IsVerification() bool
	Dirs() TempDirs
}

func New(name string, cfg Config, options ...Option) (Build, error) {
	return newCore(name, cfg, options...)
}

type core struct {
	Settings
	config Config
}

func errDirtyWorktree(dirtyFiles []string) error {
	maxDirty := 10
	if len(dirtyFiles) > maxDirty {
		dirtyFiles = dirtyFiles[:maxDirty]
		dotDotDot := fmt.Sprintf("\n%d more dirty files not shown...", len(dirtyFiles)-maxDirty)
		dirtyFiles = append(dirtyFiles, dotDotDot)
	}
	list := "\n" + strings.Join(dirtyFiles, "\n")
	return fmt.Errorf("worktree dirty but build is set to clean only; dirty paths: %s", list)
}

func newCore(name string, cfg Config, options ...Option) (*core, error) {
	s, err := newSettings(options)
	if err != nil {
		return nil, err
	}
	if s.cleanOnly && cfg.Product.IsDirty() {
		return nil, errDirtyWorktree(cfg.Product.DirtyFiles)
	}
	return &core{
		Settings: s,
		config:   cfg,
	}, nil
}

func (b *core) Config() Config {
	return b.config
}

func (b *core) IsVerification() bool { return b.isVerification }

func (b *core) Dirs() TempDirs {
	return newDirsFromConfig(b.config, b.isVerification)
}

func (b *core) ChangeRoot(dir string) error {
	b.Debug("changing root to %s", dir)
	var err error
	b.config, err = b.config.ChangeRoot(dir)
	return err
}

func (b *core) ChangeToVerificationRoot() error {
	return b.ChangeRoot(b.config.VerificationRoot())
}

func (b *core) ChangeToPrimaryRoot() error {
	return b.ChangeRoot(b.config.RemotePrimaryRoot())
}

// UpdateBuildRoot updates the build root for this build depending
// whether it's a primary or verification build.
func (b *core) UpdateBuildRoot() error {
	if b.isVerification {
		return b.ChangeToVerificationRoot()
	}
	return b.ChangeToPrimaryRoot()
}

func (b *core) CachedResult() (Result, bool, error) {
	var r Result
	path := b.config.BuildResultCachePath(b.isVerification)
	exists, err := fs.FileExists(path)
	if err != nil {
		b.Debug("Cache read error: %s", err)
		return r, false, err
	}
	if !exists {
		b.Debug("Cache miss: %s", path)
		return r, false, nil
	}
	b.Debug("Cache hit: %s", path)
	r, err = json.ReadFile[Result](path)
	r.loadedFromCache = true
	return r, err == nil, err
}

func newStep(desc string, action StepFunc) Step {
	return Step{desc: desc, action: action}
}

func (b *core) Steps() []Step {
	var productRevisionTimestamp time.Time
	return []Step{
		newStep("validating inputs", func() error {
			var err error
			productRevisionTimestamp, err = b.Config().Product.RevisionTimestamp()
			return err
		}),

		newStep("creating output directories", b.createDirectories),

		newStep("running build instructions", b.runInstructions),

		newStep("asserting executable written", b.assertExecutableWritten),

		newStep("setting mtimes", func() error {
			return fs.SetMtimes(b.Config().Paths.TargetDir(), productRevisionTimestamp)
		}),

		newStep(fmt.Sprintf("creating zip file %q", b.Config().Paths.ZipPath), func() error {
			return zipper.ZipToFile(b.Config().Paths.TargetDir(), b.Config().Paths.ZipPath, b.Settings.Log)
		}),
	}
}

func (b *core) createDirectories() error {
	c := b.config
	if err := fs.MkdirEmpty(c.Paths.TargetDir()); err != nil {
		return err
	}
	return fs.Mkdirs(c.Paths.ZipDir(), c.Paths.MetaDir)
}

func (b *core) assertExecutableWritten() error {
	cmd := exec.Command("ls", "-lR")
	dir_output, _ := cmd.Output()
	fmt.Printf("Directory info (ls -lR): %s\n", string(dir_output))
	binExists, err := b.executableWasWritten()
	if err != nil {
		return err
	}
	if !binExists {
		return fmt.Errorf("no file written to BIN_PATH %q", b.config.Paths.BinPath)
	}
	return nil
}

func (b *core) executableWasWritten() (bool, error) {
	return fs.FileExists(b.config.Paths.BinPath)
}

func (b *core) newCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(b.Settings.context, name, args...)
	cmd.Dir = b.config.Paths.WorkDir
	cmd.Stdout = b.Settings.stdout
	cmd.Stderr = b.Settings.stderr
	return cmd
}

func (b *core) runCommand(name string, args ...string) error {
	return b.newCommand(name, args...).Run()
}

func (b *core) runInstructions() error {
	path, err := b.writeInstructions()
	if err != nil {
		return err
	}

	c := b.newCommand(b.Settings.bash, path)
	c.Env = b.Env()
	b.Log("Build environment determined by config:\n%s", strings.Join(c.Env, "\n"))
	c.Env = append(os.Environ(), c.Env...)
	b.Debug("Full build environment:\n%s", strings.Join(c.Env, "\n"))

	return c.Run()
}

// writeInstructions writes the build instructions to a temporary file
// and returns its path, or an error if writing fails.
func (b *core) writeInstructions() (path string, err error) {
	b.Log("Build instructions:\n%s", b.config.Parameters.Instructions)
	return fs.WriteTempFile("actions-go-build.instructions", b.config.Parameters.Instructions)
}
