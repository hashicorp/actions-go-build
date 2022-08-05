package build

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/actions-go-build/internal/log"
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

func New(cfg Config, options ...Option) (Build, error) {
	return newCore(cfg, options...)
}

func newCore(cfg Config, options ...Option) (*core, error) {
	s, err := newSettings(options)
	if err != nil {
		return nil, err
	}
	return &core{
		settings: s,
		config:   cfg,
	}, nil
}

type core struct {
	settings      Settings
	config        Config
	prebuildSteps []Step
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

func (b *core) log(f string, a ...any) {
	b.settings.logFunc(f, a...)
}

func (b *core) CachedResult() (Result, bool, error) {
	var r Result
	path := b.config.buildResultCachePath()
	exists, err := fs.FileExists(path)
	if err != nil {
		log.Debug("Cache read error: %s", err)
		return r, false, err
	}
	if !exists {
		log.Debug("Cache miss: %s", path)
		return r, false, nil
	}
	log.Debug("Cache hit: %s", path)
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
			return zipper.ZipToFile(c.Paths.TargetDir, c.Paths.ZipPath, b.settings.logFunc)
		}),
	}
}

func (b *core) createDirectories() error {
	c := b.config
	b.log("Creating output directories.")
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
	cmd := exec.CommandContext(b.settings.context, name, args...)
	cmd.Dir = b.config.Paths.WorkDir
	cmd.Stdout = b.settings.stdout
	cmd.Stderr = b.settings.stderr
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

	b.log("Running build instructions with environment:")
	env := b.Env()
	for _, e := range b.Env() {
		b.log(e)
	}
	c := b.newCommand(b.settings.bash, path)
	c.Env = os.Environ()
	c.Env = append(c.Env, env...)
	return c.Run()
}

// writeInstructions writes the build instructions to a temporary file
// and returns its path, or an error if writing fails.
func (b *core) writeInstructions() (path string, err error) {
	b.log("Writing build instructions to temp file.")
	return fs.WriteTempFile("actions-go-build.instructions", b.config.Parameters.Instructions)
}

func (b *core) listInstructions() {
	b.log("Listing build instructions...")
	b.log(b.config.Parameters.Instructions)
}
