package config

import (
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

// ExportToGitHubEnv writes GitHub Actions set env commands to the provided writer.
// Use os.Stdout as the writer when you want GitHub to see the commands, use other
// writers for testing.
func (c Config) ExportToGitHubEnv() {
	gh := githubactions.New()
	gh.SetEnv("PRODUCT_NAME", c.ProductName)
	gh.SetEnv("PRODUCT_VERSION", c.ProductVersion)
	gh.SetEnv("PRODUCT_REVISION", c.ProductRevision)
	gh.SetEnv("PRODUCT_REVISION_TIME", c.ProductRevisionTime)
	gh.SetEnv("GO_VERSION", c.GoVersion)
	gh.SetEnv("OS", c.OS)
	gh.SetEnv("ARCH", c.Arch)
	gh.SetEnv("REPRODUCIBLE", c.Reproducible)
	gh.SetEnv("INSTRUCTIONS", c.Instructions)
	gh.SetEnv("BIN_NAME", c.BinName)
	gh.SetEnv("ZIP_NAME", c.ZipName)
	gh.SetEnv("PRIMARY_BUILD_ROOT", c.PrimaryBuildRoot)
	gh.SetEnv("VERIFICATION_BUILD_ROOT", c.VerificationBuildRoot)
	gh.SetEnv("PRIMARY_BIN_PATH", c.PrimaryBuild.BinPath)
	gh.SetEnv("PRIMARY_ZIP_PATH", c.PrimaryBuild.ZipPath)
	gh.SetEnv("VERIFICATION_BIN_PATH", c.VerificationBuild.BinPath)
	gh.SetEnv("VERIFICATION_ZIP_PATH", c.VerificationBuild.ZipPath)
	gh.SetEnv("TARGET_DIR", c.TargetDir)
	gh.SetEnv("ZIP_DIR", c.ZipDir)
	gh.SetEnv("META_DIR", c.MetaDir)
}
