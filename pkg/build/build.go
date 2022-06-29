package build

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/internal/fs"
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

func New(cfg config.Config, options ...Option) (Build, error) {
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
	config   config.Config
}

type dirs struct {
	source, target, zip, meta string
}

func (b *build) Run() error {
	c := b.config
	log.Printf("Starting build process.")
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	log.Printf("Beginning build, rooted at %q", wd)
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

	// TODO
	//   - Verify artifact written to BIN_PATH
	//   - Record bin digest.
	//   - Set mtime of all files in TARGET_DIR
	//   - Zip contents of TARGET_DIR
	//   - Record zip digest.

	return nil
}

func (b *build) runInstructions(path string) error {
	cmd := exec.CommandContext(b.settings.context, b.settings.bash, path)
	cmd.Stdout = b.settings.stdout
	cmd.Stderr = b.settings.stderr

	return cmd.Run()
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
