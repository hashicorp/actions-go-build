package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

// Inputs roughly maps to the set of action inputs.
type Inputs struct {
	// Product contains the product details.
	Product crt.Product

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

func (i Inputs) Config(rc crt.RepoContext) (Config, error) {
	i = i.init(rc)
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

func (i Inputs) init(rc crt.RepoContext) Inputs {
	return i.trimSpace().setDefaults(rc)
}

// trimSpace trims leading and trailing whitespace from every input string.
func (i Inputs) trimSpace() Inputs {
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

func (i Inputs) setDefaults(rc crt.RepoContext) Inputs {
	i.Product = i.Product.Init(rc)
	if i.BinName == "" {
		i.BinName = i.Product.Name
	}
	if i.ZipName == "" {
		i.ZipName = fmt.Sprintf("%s_%s_%s_%s.zip", i.Product.Name, i.Product.Version, i.OS, i.Arch)
	}
	if i.PrimaryBuildRoot == "" {
		i.PrimaryBuildRoot = rc.Dir
	}
	if i.VerificationBuildRoot == "" {
		i.VerificationBuildRoot = siblingPath(i.PrimaryBuildRoot, "verification")
	}
	return i
}

// siblingPath returns the sibling path name of to. Sibling path is defined as
// a path with the same directory component but a different base name.
func siblingPath(to, name string) string {
	return filepath.Join(filepath.Dir(to), name)
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
