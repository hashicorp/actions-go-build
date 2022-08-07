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
	Kind() string
	ChangeRoot(string) error
}

func New(name string, cfg Config, options ...Option) (Build, error) {
	return newCore(name, cfg, options...)
}

type core struct {
	Settings
	config Config
}

func newCore(name string, cfg Config, options ...Option) (*core, error) {
	s, err := newSettings("build: "+name, options)
	if err != nil {
		return nil, err
	}
	s.Debug("Initialised")
	return &core{
		Settings: s,
		config:   cfg,
	}, nil
}

func (b *core) Config() Config {
	return b.config
}

func (b *core) Kind() string { return "unknown" }

func (b *core) ChangeRoot(dir string) error {
	var err error
	b.config, err = b.config.ChangeRoot(dir)
	return err
}

func (b *core) CachedResult() (Result, bool, error) {
	var r Result
	path := b.config.buildResultCachePath()
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
	return r, err == nil, err
}

func newStep(desc string, action StepFunc) Step {
	return Step{desc: desc, action: action}
}

func (b *core) Steps() []Step {
	c := b.config
	var productRevisionTimestamp time.Time

	return []Step{
		newStep("validating inputs", func() error {
			var err error
			productRevisionTimestamp, err = c.Product.RevisionTimestamp()
			return err
		}),

		newStep("creating output directories", b.createDirectories),

		newStep("running build instructions", b.runInstructions),

		newStep("asserting executable written", b.assertExecutableWritten),

		newStep("setting mtimes", func() error {
			return fs.SetMtimes(c.Paths.TargetDir, productRevisionTimestamp)
		}),

		newStep("creating zip file", func() error {
			return zipper.ZipToFile(c.Paths.TargetDir, c.Paths.ZipPath, b.Settings.Log)
		}),
	}
}

func (b *core) createDirectories() error {
	c := b.config
	b.Log("Creating output directories.")
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

	b.Log("Running build instructions with environment:")
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
	b.Log("Writing build instructions to temp file.")
	return fs.WriteTempFile("actions-go-build.instructions", b.config.Parameters.Instructions)
}

func (b *core) listInstructions() {
	b.Log("Listing build instructions...")
	b.Log(b.config.Parameters.Instructions)
}
