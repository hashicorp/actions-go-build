package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/internal/zipper"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
)

type Build interface {
	Run() error
	Env() []string
}

func resolveBashPath(path string) (string, error) {
	if path == "" {
		path = "bash"
	}
	return exec.LookPath(path)
}

func New(cfg crt.BuildConfig, options ...Option) (Build, error) {
	s, err := newSettings(options)
	if err != nil {
		return &build{}, err
	}
	return &build{
		settings: s,
		config:   cfg,
	}, nil
}

type build struct {
	settings Settings
	config   crt.BuildConfig
}

type dirs struct {
	source, target, zip, meta string
}

func (b *build) Run() error {
	c := b.config
	log.Printf("Starting build process.")

	// A quick bit of validation before running the build.
	productRevisionTimestamp, err := c.Product.RevisionTimestamp()
	if err != nil {
		return err
	}

	log.Printf("Beginning build, rooted at %q", b.config.Paths.WorkDir)
	if err := fs.Mkdirs(c.Paths.TargetDir, c.Paths.ZipDir(), c.Paths.MetaDir); err != nil {
		return err
	}
	instructionsPath, err := b.writeInstructions()
	if err != nil {
		return err
	}

	b.listInstructions()

	if err := b.runInstructions(instructionsPath); err != nil {
		return err
	}

	binExists, err := fs.FileExists(c.Paths.BinPath)
	if err != nil {
		return err
	}
	if !binExists {
		return fmt.Errorf("no file written to BIN_PATH %q", c.Paths.BinPath)
	}

	if err := b.writeDigest(c.Paths.BinPath, "bin_digest"); err != nil {
		return err
	}

	if err := fs.SetMtimes(c.Paths.TargetDir, productRevisionTimestamp); err != nil {
		return err
	}

	if err := zipper.ZipToFile(c.Paths.TargetDir, c.Paths.ZipPath); err != nil {
		return err
	}

	if err := b.writeDigest(c.Paths.ZipPath, "zip_digest"); err != nil {
		return err
	}

	return nil
}

func (b *build) writeDigest(of, named string) error {
	sha, err := digest.FileSHA256Hex(of)
	if err != nil {
		return err
	}

	digestPath := filepath.Join(b.config.Paths.MetaDir, named)

	return fs.WriteFile(digestPath, sha)
}

func (b *build) newCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(b.settings.context, name, args...)
	cmd.Dir = b.config.Paths.WorkDir
	cmd.Stdout = b.settings.stdout
	cmd.Stderr = b.settings.stderr
	return cmd
}

func (b *build) runCommand(name string, args ...string) error {
	return b.newCommand(name, args...).Run()
}

func (b *build) runInstructions(path string) error {
	log.Printf("Running build instructions with environment:")
	env := b.Env()
	for _, e := range b.Env() {
		fmt.Fprintln(b.settings.stderr, e)
	}
	c := b.newCommand(b.settings.bash, path)
	c.Env = os.Environ()
	c.Env = append(c.Env, env...)
	return c.Run()
}

// writeInstructions writes the build instructions to a temporary file
// and returns its path, or an error if writing fails.
func (b *build) writeInstructions() (path string, err error) {
	log.Printf("Writing build instructions to temp file.")
	return fs.WriteTempFile("actions-go-build.instructions", b.config.Parameters.Instructions)
}

func (b *build) listInstructions() {
	log.Printf("Listing build instructions...")
	log.Println(b.config.Parameters.Instructions)
}
