package config

import (
	"log"
	"os"
	"path/filepath"

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
	ProductRevision     string
	ProductRevisionTime string
	PrimaryBuild        BuildConfig
	VerificationBuild   BuildConfig
	ProductCoreName     string
	TargetDir           string
	ZipDir              string
	MetaDir             string
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
func (c Config) ExportToGitHubEnv() {
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
	es.setEnv("ZIP_NAME", c.ZipName)
	es.setEnv("PRIMARY_BUILD_ROOT", c.PrimaryBuildRoot)
	es.setEnv("VERIFICATION_BUILD_ROOT", c.VerificationBuildRoot)
	es.setEnv("PRIMARY_BIN_PATH", c.PrimaryBuild.BinPath)
	es.setEnv("PRIMARY_ZIP_PATH", c.PrimaryBuild.ZipPath)
	es.setEnv("VERIFICATION_BIN_PATH", c.VerificationBuild.BinPath)
	es.setEnv("VERIFICATION_ZIP_PATH", c.VerificationBuild.ZipPath)
	es.setEnv("TARGET_DIR", c.TargetDir)
	es.setEnv("ZIP_DIR", c.ZipDir)
	es.setEnv("META_DIR", c.MetaDir)
}

func (es envSetter) setEnv(name, value string) {
	log.Printf("Setting %q to %q", name, value)
	if os.Getenv("GITHUB_ENV") != "" {
		es.setEnvFunc(name, value)
	}
}
