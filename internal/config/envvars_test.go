package config

import (
	"os"
	"testing"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/assert"
	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/goldenfile"
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
		want   build.Config
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
			testBuildConfig(func(bc *build.Config) {
				bc.Paths.WorkDir = "/blah"
				bc.Paths.TargetDir = "/blah/dist"
				bc.Paths.BinPath = "/blah/dist/lockbox"
				bc.Paths.ZipPath = "/blah/out/lockbox_1.2.3_linux_amd64.zip"
				bc.Paths.MetaDir = "/blah/meta"
			}),
		},
		{
			"root/blah (overridden zip name)",
			testConfig(func(c *Config) {
				c.ZipName = "blargle.zip"
			}),
			"/blah",
			testBuildConfig(func(bc *build.Config) {
				bc.Paths.WorkDir = "/blah"
				bc.Paths.TargetDir = "/blah/dist"
				bc.Paths.BinPath = "/blah/dist/lockbox"
				bc.Paths.ZipPath = "/blah/out/blargle.zip"
				bc.Paths.MetaDir = "/blah/meta"
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

func standardBuildconfig() build.Config {
	return build.Config{
		Product:    standardProduct(),
		Parameters: standardParameters(),
		Paths: build.Paths{
			WorkDir:   "/",
			TargetDir: "/dist",
			BinPath:   "/dist/lockbox",
			ZipPath:   "/out/lockbox_1.2.3_linux_amd64.zip",
			MetaDir:   "/meta",
		},
		Tool: crt.Tool{
			Name:     "thisaction",
			Version:  "0.0.0",
			Revision: "cabba9e",
		},
	}
}

func testBuildConfig(modifiers ...func(*build.Config)) build.Config {
	return applyModifiers(standardBuildconfig(), modifiers...)
}
