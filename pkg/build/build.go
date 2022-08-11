package build

import (
	"fmt"
	"os"
	"os/exec"
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
}

func New(name string, isVerification bool, cfg Config, options ...Option) (Build, error) {
	return newCore(name, isVerification, cfg, options...)
}

type core struct {
	Settings
	config         Config
	isVerification bool
}

func newCore(name string, isVerification bool, cfg Config, options ...Option) (*core, error) {
	s, err := newSettings(options)
	if err != nil {
		return nil, err
	}
	return &core{
		Settings:       s,
		config:         cfg,
		isVerification: isVerification,
	}, nil
}

func (b *core) Config() Config {
	return b.config
}

func (b *core) IsVerification() bool { return b.isVerification }

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

func (b *core) CachedResult() (Result, bool, error) {
	var r Result
	path := b.config.buildResultCachePath(b.isVerification)
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
			return fs.SetMtimes(b.Config().Paths.TargetDir, productRevisionTimestamp)
		}),

		newStep(fmt.Sprintf("creating zip file %q", b.Config().Paths.ZipPath), func() error {
			return zipper.ZipToFile(b.Config().Paths.TargetDir, b.Config().Paths.ZipPath, b.Settings.Log)
		}),
	}
}

func (b *core) createDirectories() error {
	c := b.config
	return fs.Mkdirs(c.Paths.TargetDir, c.Paths.ZipDir(), c.Paths.MetaDir)
}

func (b *core) assertExecutableWritten() error {
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

	b.listInstructions()

	b.Log("build environment:")
	env := b.Env()
	for _, e := range b.Env() {
		b.Log(e)
	}
	c := b.newCommand(b.Settings.bash, path)
	c.Env = os.Environ()
	c.Env = append(c.Env, env...)
	return c.Run()
}

// writeInstructions writes the build instructions to a temporary file
// and returns its path, or an error if writing fails.
func (b *core) writeInstructions() (path string, err error) {
	return fs.WriteTempFile("actions-go-build.instructions", b.config.Parameters.Instructions)
}

func (b *core) listInstructions() {
	b.Log(b.config.Parameters.Instructions)
}
