package config

import (
	"os"
	"testing"

	"github.com/hashicorp/actions-go-build/internal/testhelpers/assert"
	"github.com/hashicorp/actions-go-build/internal/testhelpers/goldenfile"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

func TestConfig_ExportToGitHubEnv_ok(t *testing.T) {
	goldenfile.Do(t, func(got *os.File) {
		os.Setenv("GITHUB_ENV", got.Name())
		c := standardConfig()
		c.ExportToGitHubEnv()
	})
}

func TestConfig_BuildConfig_ok(t *testing.T) {
	cases := []struct {
		desc   string
		config Config
		root   string
		want   crt.BuildConfig
	}{
		{
			"root",
			testConfig(),
			"/",
			testBuildConfig(),
		},
		{
			"root/blah",
			testConfig(),
			"/blah",
			testBuildConfig(func(bc *crt.BuildConfig) {
				bc.WorkDir = "/blah"
				bc.TargetDir = "/blah/dist"
				bc.BinPath = "/blah/dist/lockbox"
				bc.ZipPath = "/blah/out/lockbox_1.2.3_linux_amd64.zip"
				bc.ZipDir = "/blah/out"
				bc.MetaDir = "/blah/meta"
			}),
		},
		{
			"root/blah (overridden zip name)",
			testConfig(func(c *Config) {
				c.ZipName = "blargle.zip"
			}),
			"/blah",
			testBuildConfig(func(bc *crt.BuildConfig) {
				bc.WorkDir = "/blah"
				bc.TargetDir = "/blah/dist"
				bc.BinPath = "/blah/dist/lockbox"
				bc.ZipPath = "/blah/out/blargle.zip"
				bc.ZipDir = "/blah/out"
				bc.MetaDir = "/blah/meta"
			}),
		},
	}

	for _, c := range cases {
		desc, config, root, want := c.desc, c.config, c.root, c.want
		t.Run(desc, func(t *testing.T) {
			got, err := config.buildConfig(root)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, got, want)
		})
	}
}

func standardBuildconfig() crt.BuildConfig {
	return crt.BuildConfig{
		Product:      standardProduct(),
		TargetOS:     "linux",
		TargetArch:   "amd64",
		WorkDir:      "/",
		TargetDir:    "/dist",
		BinPath:      "/dist/lockbox",
		ZipPath:      "/out/lockbox_1.2.3_linux_amd64.zip",
		Instructions: `go build -o "$BIN_PATH"`,
		ZipDir:       "/out",
		MetaDir:      "/meta",
	}
}

func testBuildConfig(modifiers ...func(*crt.BuildConfig)) crt.BuildConfig {
	return applyModifiers(standardBuildconfig(), modifiers...)
}
