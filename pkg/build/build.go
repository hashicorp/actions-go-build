package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/internal/fs"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
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
	log.Printf("Beginning build, rooted at %q", b.config.WorkDir)
	if err := fs.Mkdirs(c.TargetDir, c.ZipDir, c.MetaDir); err != nil {
		return err
	}
	instructionsPath, err := b.writeInstructions()
	if err != nil {
		return err
	}

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
