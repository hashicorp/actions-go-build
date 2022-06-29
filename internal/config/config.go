package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sethvargo/go-githubactions"
)

// BuildConfig contains the
type BuildConfig struct {
	// WorkDir is the absolute directory to run the build instructions in.
	WorkDir string
	// TargetDir is the absolute path to the dir where any other files
	// needed to be included in the zip file should be placed.
	TargetDir string
	// BinPath is the path to the executable binary the instructions must create.
	BinPath string
	// ZipPath is the path to the zip file that will be created.
	ZipPath string
}

// Config is a complete configuration for this action.
type Config struct {
	Inputs
	ProductRevision     string
	ProductRevisionTime string
	ProductCoreName     string
	TargetDir           string
	ZipDir              string
	MetaDir             string
}

func (c Config) BuildConfig(root string) (BuildConfig, error) {
	if !filepath.IsAbs(root) {
		return BuildConfig{}, fmt.Errorf("root path %q is not absolute", root)
	}
	return BuildConfig{
		WorkDir:   root,
		TargetDir: filepath.Join(root, c.TargetDir),
		BinPath:   filepath.Join(root, c.TargetDir, c.BinName),
		ZipPath:   filepath.Join(root, c.ZipDir, c.ZipName),
	}, nil
}

func (c Config) PrimaryBuildConfig() (BuildConfig, error) {
	return c.BuildConfig(c.PrimaryBuildRoot)
}

func (c Config) VerificationBuildConfig() (BuildConfig, error) {
	return c.BuildConfig(c.VerificationBuildRoot)
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

// ExportToGitHubEnv writes GitHub Actions set env commands to the provided writer.
// Use os.Stdout as the writer when you want GitHub to see the commands, use other
// writers for testing.
func (c Config) ExportToGitHubEnv() error {
	primary, err := c.PrimaryBuildConfig()
	if err != nil {
		return err
	}
	verification, err := c.VerificationBuildConfig()
	if err != nil {
		return err
	}

	es := newEnvSetter()
	es.setEnv("PRODUCT_NAME", c.ProductName)
	es.setEnv("PRODUCT_VERSION", c.ProductVersion)
	es.setEnv("PRODUCT_REVISION", c.ProductRevision)
	es.setEnv("PRODUCT_REVISION_TIME", c.ProductRevisionTime)
	es.setEnv("GO_VERSION", c.GoVersion)
	es.setEnv("OS", c.OS)
	es.setEnv("ARCH", c.Arch)
	es.setEnv("REPRODUCIBLE", c.Reproducible)
	es.setEnv("INSTRUCTIONS", c.Instructions)
	es.setEnv("BIN_NAME", c.BinName)
	es.setEnv("BIN_PATH", filepath.Join(c.TargetDir, c.BinName))
	es.setEnv("ZIP_PATH", filepath.Join(c.ZipDir, c.ZipName))
	es.setEnv("ZIP_NAME", c.ZipName)
	es.setEnv("PRIMARY_BUILD_ROOT", c.PrimaryBuildRoot)
	es.setEnv("VERIFICATION_BUILD_ROOT", c.VerificationBuildRoot)
	es.setEnv("BIN_PATH_PRIMARY", primary.BinPath)
	es.setEnv("ZIP_PATH_PRIMARY", primary.ZipPath)
	es.setEnv("BIN_PATH_VERIFICATION", verification.BinPath)
	es.setEnv("ZIP_PATH_VERIFICATION", verification.ZipPath)
	es.setEnv("TARGET_DIR", c.TargetDir)
	es.setEnv("ZIP_DIR", c.ZipDir)
	es.setEnv("META_DIR", c.MetaDir)

	// Extra vars set for the build environment.
	es.setEnv("GOOS", c.OS)
	es.setEnv("GOARCH", c.Arch)

	return nil
}

func (es envSetter) setEnv(name, value string) {
	log.Printf("Setting %q to %q", name, value)
	if os.Getenv("GITHUB_ENV") != "" {
		es.setEnvFunc(name, value)
	}
}
