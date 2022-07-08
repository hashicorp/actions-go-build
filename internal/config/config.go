package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/sethvargo/go-envconfig"
	"github.com/sethvargo/go-githubactions"
)

// BuildConfig contains the
type BuildConfig struct {
	BinPath string
	ZipPath string
}

// Config is a complete configuration for this action.
type Config struct {
	Inputs
	TargetDir string
	ZipDir    string
	MetaDir   string
}

// FromEnvironment creates a new Config from environment variables
// and repository context in the current working directory.
func FromEnvironment() (Config, error) {
	var inputs Inputs
	ctx := context.Background()
	if err := envconfig.Process(ctx, &inputs); err != nil {
		return Config{}, err
	}

	wd, err := os.Getwd()
	if err != nil {
		return Config{}, err
	}

	bc, err := crt.GetRepoContext(wd)
	if err != nil {
		return Config{}, err
	}

	return inputs.Config(bc)
}

// buildConfig returns a BuildConfig based on this Config, rooted at root.
// The root must be an absolute path.
func (c Config) buildConfig(root string) (crt.BuildConfig, error) {
	if !filepath.IsAbs(root) {
		return crt.BuildConfig{}, fmt.Errorf("root path %q is not absolute", root)
	}
	return crt.BuildConfig{
		Product:            c.Product,
		ProductVersionMeta: "",
		WorkDir:            root,
		TargetDir:          filepath.Join(root, c.TargetDir),
		BinPath:            filepath.Join(root, c.TargetDir, c.BinName),
		ZipPath:            filepath.Join(root, c.ZipDir, c.ZipName),
		Instructions:       c.Instructions,
		TargetOS:           c.OS,
		TargetArch:         c.Arch,
		ZipDir:             filepath.Join(root, c.ZipDir),
		MetaDir:            filepath.Join(root, c.MetaDir),
	}, nil
}

// PrimaryBuildConfig returns the config for the primary build.
func (c Config) PrimaryBuildConfig() (crt.BuildConfig, error) {
	return c.buildConfig(c.PrimaryBuildRoot)
}

// VerificationBuildConfig returns the config for the verification build.
func (c Config) VerificationBuildConfig() (crt.BuildConfig, error) {
	return c.buildConfig(c.VerificationBuildRoot)
}

type envSetter struct {
	setEnvFunc func(name, value string)
}

func newEnvSetter() envSetter {
	if os.Getenv("GITHUB_ENV") != "" {
		return envSetter{githubactions.SetEnv}
	}
	log.Printf("WARNING: GITHUB_ENV not set, just printing environment.")
	return envSetter{nil}
}

type EnvVar struct{ Name, Value string }

func (c Config) EnvVars() ([]EnvVar, error) {
	var kvs []EnvVar
	addEnv := func(key, value string) {
		kvs = append(kvs, EnvVar{key, value})
	}

	// TODO don't serialise primary and verification build configs to env here.
	// We can derive them from the rest of the config anyway so there's probably
	// no point writing them to GITHUB_ENV.
	//
	// Keeping them here for now since the current bash implementation expects
	// to see them.

	primary, err := c.PrimaryBuildConfig()
	if err != nil {
		return nil, err
	}
	verification, err := c.VerificationBuildConfig()
	if err != nil {
		return nil, err
	}

	addEnv("PRODUCT_NAME", c.Product.Name)
	addEnv("PRODUCT_VERSION", c.Product.Version)
	addEnv("PRODUCT_REVISION", c.Product.Revision)
	addEnv("PRODUCT_REVISION_TIME", c.Product.RevisionTime)
	addEnv("GO_VERSION", c.GoVersion)
	addEnv("OS", c.OS)
	addEnv("ARCH", c.Arch)
	addEnv("REPRODUCIBLE", c.Reproducible)
	addEnv("INSTRUCTIONS", c.Instructions)
	addEnv("BIN_NAME", c.BinName)
	addEnv("BIN_PATH", filepath.Join(c.TargetDir, c.BinName))
	addEnv("ZIP_PATH", filepath.Join(c.ZipDir, c.ZipName))
	addEnv("ZIP_NAME", c.ZipName)
	addEnv("PRIMARY_BUILD_ROOT", c.PrimaryBuildRoot)
	addEnv("VERIFICATION_BUILD_ROOT", c.VerificationBuildRoot)
	addEnv("BIN_PATH_PRIMARY", primary.BinPath)
	addEnv("ZIP_PATH_PRIMARY", primary.ZipPath)
	addEnv("BIN_PATH_VERIFICATION", verification.BinPath)
	addEnv("ZIP_PATH_VERIFICATION", verification.ZipPath)
	addEnv("TARGET_DIR", c.TargetDir)
	addEnv("ZIP_DIR", c.ZipDir)
	addEnv("META_DIR", c.MetaDir)

	return kvs, nil
}

func (c Config) foreach(fn func(key, value string)) error {
	vars, err := c.EnvVars()
	if err != nil {
		return err
	}
	for _, pair := range vars {
		fn(pair.Name, pair.Value)
	}
	return nil
}

// ExportToGitHubEnv writes this config to GITHUB_ENV so it can be read by
// future steps in this job. If GITHUB_ENV isn't set, it prints a warning
// and just logs what would have been set.
func (c Config) ExportToGitHubEnv() error {
	if os.Getenv("GITHUB_ENV") == "" {
		return errors.New("GITHUB_ENV not set")
	}
	es := newEnvSetter()
	return c.foreach(es.setEnv)
}

func (es envSetter) setEnv(name, value string) {
	log.Printf("Setting %q to %q", name, value)
	es.setEnvFunc(name, value)
}
