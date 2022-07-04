package build

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/internal/fs"
	"github.com/hashicorp/actions-go-build/pkg/digest"
)

type Build interface {
	Run() error
}

// Settings contains settings for running the instructions.
// Don't use this directly, use the With... functions to set
// settings when calling New.
type Settings struct {
	bash    string
	context context.Context
	stdout  io.Writer
	stderr  io.Writer
}

func resolveBashPath(path string) (string, error) {
	if path == "" {
		path = "bash"
	}
	return exec.LookPath(path)
}

func (s *Settings) setDefaults() (err error) {
	s.bash, err = resolveBashPath(s.bash)
	if err != nil {
		return err
	}
	if s.context == nil {
		s.context = context.Background()
	}
	if s.stdout == nil {
		s.stdout = os.Stdout
	}
	if s.stderr == nil {
		s.stderr = os.Stderr
	}
	return nil
}

type Option func(*Settings)

func WithContext(c context.Context) Option { return func(s *Settings) { s.context = c } }
func WithStdout(w io.Writer) Option        { return func(s *Settings) { s.stdout = w } }
func WithStderr(w io.Writer) Option        { return func(s *Settings) { s.stderr = w } }
func WithBash(path string) Option          { return func(s *Settings) { s.bash = path } }

func New(cfg config.BuildConfig, options ...Option) (Build, error) {
	s := &Settings{}
	for _, option := range options {
		option(s)
	}
	if err := s.setDefaults(); err != nil {
		return nil, err
	}
	return &build{
		settings: *s,
		config:   cfg,
	}, nil
}

type build struct {
	settings Settings
	config   config.BuildConfig
}

type dirs struct {
	source, target, zip, meta string
}

func (b *build) Run() error {
	c := b.config
	log.Printf("Starting build process.")
	log.Printf("Beginning build, rooted at %q", b.config.WorkDir)
	if err := fs.Mkdirs(c.TargetDir, c.ZipDir, c.MetaDir); err != nil {
		return err
	}
	instructionsPath, err := b.writeInstructions()
	if err != nil {
		return err
	}

	log.Printf("Running build instructions...")

	if err := b.runInstructions(instructionsPath); err != nil {
		return err
	}

	binExists, err := fs.FileExists(c.BinPath)
	if err != nil {
		return err
	}
	if !binExists {
		return fmt.Errorf("no file written to BIN_PATH %q", c.BinPath)
	}

	binSHA, err := digest.FileSHA256Hex(c.BinPath)
	if err != nil {
		return err
	}

	binDigestPath := filepath.Join(c.MetaDir, "bin_digest")

	if err := fs.WriteFile(binDigestPath, binSHA); err != nil {
		return err
	}

	// TODO
	//   - Set mtime of all files in TARGET_DIR
	//   - Zip contents of TARGET_DIR
	//   - Record zip digest.

	return nil
}

func (b *build) newCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(b.settings.context, name, args...)
	cmd.Dir = b.config.WorkDir
	cmd.Stdout = b.settings.stdout
	cmd.Stderr = b.settings.stderr
	return cmd
}

func (b *build) runCommand(name string, args ...string) error {
	return b.newCommand(name, args...).Run()
}

func (b *build) runInstructions(path string) error {
	c := b.newCommand(b.settings.bash, path)
	c.Env = os.Environ()
	c.Env = append(c.Env, b.buildEnv()...)
	return c.Run()
}

type envVar struct {
	name, desc string
	valueFunc  func(config.BuildConfig) string
}

func (ev envVar) String(c config.BuildConfig) string {
	return fmt.Sprintf("%s=%s", ev.name, ev.valueFunc(c))
}

func buildEnvDef() []envVar {
	return []envVar{
		{
			"TARGET_DIR",
			"Absolute path to the zip contents directory.",
			func(c config.BuildConfig) string { return c.TargetDir },
		},
		{
			"PRODUCT_NAME",
			"Same as the `product_name` input.",
			func(c config.BuildConfig) string { return c.Product.Name },
		},
		{
			"PRODUCT_VERSION",
			"Same as the `product_version` input.",
			func(c config.BuildConfig) string { return c.Product.Version },
		},
		{
			"PRODUCT_REVISION",
			"The git commit SHA of the product repo being built.",
			func(c config.BuildConfig) string { return c.Product.Revision },
		},
		{
			"PRODUCT_REVISION_TIME",
			"UTC timestamp of the `PRODUCT_REVISION` commit in iso-8601 format.",
			func(c config.BuildConfig) string { return c.Product.RevisionTime },
		},
		// NOTE omitting BIN_NAME as not strictly needed.
		{
			"BIN_PATH",
			"Absolute path to where instructions must write Go executable.",
			func(c config.BuildConfig) string { return c.BinPath },
		},
		{
			"OS",
			"Same as the `os` input.",
			func(c config.BuildConfig) string { return c.TargetOS },
		},
		{
			"ARCH",
			"Same as the `arch` input.",
			func(c config.BuildConfig) string { return c.TargetArch },
		},
		{
			"GOOS",
			"Same as `OS`.",
			func(c config.BuildConfig) string { return c.TargetOS },
		},
		{
			"GOARCH",
			"Same as `ARCH`.",
			func(c config.BuildConfig) string { return c.TargetArch },
		},
	}
}

func (b *build) buildEnv() []string {
	bed := buildEnvDef()
	env := make([]string, len(bed))
	for i, e := range buildEnvDef() {
		env[i] = e.String(b.config)
	}
	return env
}

func (b *build) writeInstructions() (path string, err error) {
	c := b.config
	log.Printf("Writing build instructions to temp file.")
	tempFile, err := os.CreateTemp("", "instructions.*")
	if err != nil {
		return "", err
	}
	defer func() {
		err = tempFile.Close()
	}()
	if _, err := tempFile.WriteString(c.Instructions); err != nil {
		return "", err
	}
	log.Printf("Listing build instructions...")
	log.Println(c.Instructions)
	return tempFile.Name(), nil
}
