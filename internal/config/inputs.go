package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// Inputs roughly maps to the set of action inputs.
type Inputs struct {
	// ProductRepository isn't really an input, but it's
	// required for creating default values.
	ProductRepository string `env:"PRODUCT_REPOSITORY"`

	// Required inputs.
	ProductName    string `env:"PRODUCT_NAME"`
	ProductVersion string `env:"PRODUCT_VERSION"`
	GoVersion      string `env:"GO_VERSION"`
	OS             string `env:"OS"`
	Arch           string `env:"ARCH"`
	Reproducible   string `env:"REPRODUCIBLE"`
	Instructions   string `env:"INSTRUCTIONS"`

	// Optional inputs.
	BinName string `env:"BIN_NAME"`
	ZipName string `env:"ZIP_NAME"`

	// Inputs only used in testing.
	PrimaryBuildRoot      string `env:"PRIMARY_BUILD_ROOT"`
	VerificationBuildRoot string `env:"VERIFICATION_BUILD_ROOT"`
}

type RepoContext struct {
	WorkDir    string
	CommitSHA  string
	CommitTime time.Time
}

func (i Inputs) Config(rc RepoContext) (Config, error) {
	i = i.trimSpace().setDefaults(rc)
	if err := i.validate(); err != nil {
		return Config{}, err
	}
	return Config{
		Inputs:            i,
		PrimaryBuild:      newBuildConfig(i.PrimaryBuildRoot, i.BinName, i.ZipName),
		VerificationBuild: newBuildConfig(i.VerificationBuildRoot, i.BinName, i.ZipName),
	}, nil
}

// FromEnvironment creates a new Config from environment variables.
// See Inputs for a list of the environment variables read by Digest.
func FromEnvironment(rc RepoContext) (Config, error) {
	ctx := context.Background()

	var inputs Inputs
	if err := envconfig.Process(ctx, &inputs); err != nil {
		return Config{}, err
	}

	return inputs.Config(rc)
}

// trimSpace trims leading and trailing whitespace from every input string.
func (i Inputs) trimSpace() Inputs {
	i.ProductRepository = strings.TrimSpace(i.ProductRepository)
	i.ProductName = strings.TrimSpace(i.ProductName)
	i.ProductVersion = strings.TrimSpace(i.ProductVersion)
	i.GoVersion = strings.TrimSpace(i.GoVersion)
	i.OS = strings.TrimSpace(i.OS)
	i.Arch = strings.TrimSpace(i.Arch)
	i.Reproducible = strings.TrimSpace(i.Reproducible)
	i.Instructions = strings.TrimSpace(i.Instructions)
	i.BinName = strings.TrimSpace(i.BinName)
	i.ZipName = strings.TrimSpace(i.ZipName)
	i.PrimaryBuildRoot = strings.TrimSpace(i.PrimaryBuildRoot)
	i.VerificationBuildRoot = strings.TrimSpace(i.VerificationBuildRoot)
	return i
}

func (i Inputs) setDefaults(rc RepoContext) Inputs {
	if i.ProductName == "" {
		i.ProductName = filepath.Base(i.ProductRepository)
	}
	if i.BinName == "" {
		i.BinName = i.ProductName
	}
	if i.ZipName == "" {
		i.ZipName = fmt.Sprintf("%s_%s_%s_%s.zip", i.ProductName, i.ProductVersion, i.OS, i.Arch)
	}
	if i.PrimaryBuildRoot == "" {
		i.PrimaryBuildRoot = rc.WorkDir
	}
	if i.VerificationBuildRoot == "" {
		i.VerificationBuildRoot = adjacentPath(i.PrimaryBuildRoot, "verification")
	}
	return i
}

type BuildConfig struct {
	TargetDir string
	ZipDir    string
	MetaDir   string
	BinPath   string
	ZipPath   string
}

type Config struct {
	Inputs
	PrimaryBuild      BuildConfig
	VerificationBuild BuildConfig
	ProductCoreName   string
}

func adjacentPath(to, name string) string {
	return filepath.Join(filepath.Dir(to), name)
}

func newBuildConfig(basePath, binName, zipName string) BuildConfig {
	return BuildConfig{
		TargetDir: "dist",
		ZipDir:    "out",
		MetaDir:   "meta",
		BinPath:   filepath.Join(basePath, "dist", binName),
		ZipPath:   filepath.Join(basePath, "out", zipName),
	}
}

func errRequiredInputEmpty(name string) error {
	return fmt.Errorf("required input '%s' is empty", name)
}

func (i Inputs) validate() error {
	if i.OS == "" {
		return errRequiredInputEmpty("os")
	}
	if i.Arch == "" {
		return errRequiredInputEmpty("arch")
	}
	if i.Reproducible == "" {
		return errRequiredInputEmpty("reproducible")
	}
	if i.Instructions == "" {
		return errRequiredInputEmpty("instructions")
	}
	return nil
}
