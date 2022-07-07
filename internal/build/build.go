package build

import (
	"log"
	"os"

	"github.com/hashicorp/actions-go-build/internal/fs"
)

type Build interface {
	Run() error
}

type Config struct {
	TargetDir string `env:"TARGET_DIR,required"`
	ZipDir    string `env:"ZIP_DIR,required"`
	MetaDir   string `env:"META_DIR,required"`
}

func New() Build {
	return &build{}
}

type build struct {
	dir          dirs
	instructions string
}

type dirs struct {
	source, target, zip, meta string
}

func (b *build) Run() error {
	log.Printf("Starting build process.")
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	log.Printf("Beginning build, rooted at %q", wd)
	if err := fs.Mkdirs(b.dir.target, b.dir.zip, b.dir.meta); err != nil {
		return err
	}
	return nil
}

func (b *build) writeInstructions() (path string, err error) {
	log.Printf("Writing build instructions to temp file.")
	tempFile, err := os.CreateTemp("", "instructions.*")
	if err != nil {
		return "", err
	}
	defer func() {
		err = tempFile.Close()
	}()
	if _, err := tempFile.WriteString(b.instructions); err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}
