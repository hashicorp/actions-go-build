package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sethvargo/go-envconfig"
)

// Inputs roughly maps to the set of action inputs.
type Inputs struct {
	// Product contains the product details.
	Product Product

	GoVersion    string `env:"GO_VERSION"`
	OS           string `env:"OS"`
	Arch         string `env:"ARCH"`
	Reproducible string `env:"REPRODUCIBLE"`
	Instructions string `env:"INSTRUCTIONS"`

	// Optional inputs.
	BinName string `env:"BIN_NAME"`
	ZipName string `env:"ZIP_NAME"`

	// Inputs only used in testing.
	PrimaryBuildRoot      string `env:"PRIMARY_BUILD_ROOT"`
	VerificationBuildRoot string `env:"VERIFICATION_BUILD_ROOT"`
}

type dirNames struct {
	target, zip, meta string
}

var dirs = dirNames{"dist", "out", "meta"}

func (i Inputs) Config(rc RepoContext) (Config, error) {
	i = i.trimSpace().setDefaults(rc)
	if err := i.validate(); err != nil {
		return Config{}, err
	}
	return Config{
		Inputs:    i,
		TargetDir: dirs.target,
		ZipDir:    dirs.zip,
		MetaDir:   dirs.meta,
	}, nil
}

// FromEnvironment creates a new Config from environment variables.
// See Inputs for a list of the environment variables read by Digest.
func FromEnvironment() (Config, error) {
	ctx := context.Background()

	rc, err := readRepoContext()
	if err != nil {
		return Config{}, err
	}

	var inputs Inputs
	if err := envconfig.Process(ctx, &inputs); err != nil {
		return Config{}, err
	}

	return inputs.Config(rc)
}

// trimSpace trims leading and trailing whitespace from every input string.
func (i Inputs) trimSpace() Inputs {
	i.Product = i.Product.trimSpace()
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
	i.Product = i.Product.setDefaults(rc)
	if i.BinName == "" {
		i.BinName = i.Product.Name
	}
	if i.ZipName == "" {
		i.ZipName = fmt.Sprintf("%s_%s_%s_%s.zip", i.Product.Name, i.Product.Version, i.OS, i.Arch)
	}
	if i.PrimaryBuildRoot == "" {
		i.PrimaryBuildRoot = rc.WorkDir
	}
	if i.VerificationBuildRoot == "" {
		i.VerificationBuildRoot = adjacentPath(i.PrimaryBuildRoot, "verification")
	}
	return i
}

func adjacentPath(to, name string) string {
	return filepath.Join(filepath.Dir(to), name)
}

func newBuildConfig(dirs dirNames, basePath, binName, zipName string) BuildConfig {
	return BuildConfig{
		BinPath: filepath.Join(basePath, dirs.target, binName),
		ZipPath: filepath.Join(basePath, dirs.zip, zipName),
	}
}

func errRequiredInputEmpty(name string) error {
	return fmt.Errorf("required input '%s' is empty", name)
}

// validate directly validates fields inside Product as well, because this validation
// is about the set of inputs as given, with the expectation that missing fields will
// be filled in automatically.
func (i Inputs) validate() error {
	if i.Product.Version == "" {
		return errRequiredInputEmpty("product_version")
	}
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
