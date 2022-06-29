package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/actions-go-build/internal/testhelpers/goldenfile"
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
		want   BuildConfig
	}{
		{
			"root",
			testConfig(),
			"/",
			BuildConfig{
				WorkDir:   "/",
				TargetDir: "/dist",
				BinPath:   "/dist/lockbox",
				ZipPath:   "/out/lockbox_1.2.3_linux_amd64.zip",
			},
		},
		{
			"root/blah",
			testConfig(),
			"/blah",
			BuildConfig{
				WorkDir:   "/blah",
				TargetDir: "/blah/dist",
				BinPath:   "/blah/dist/lockbox",
				ZipPath:   "/blah/out/lockbox_1.2.3_linux_amd64.zip",
			},
		},
		{
			"root/blah+ent",
			testConfig(func(c *Config) {
				c.ZipName = "blargle.zip"
			}),
			"/blah",
			BuildConfig{
				WorkDir:   "/blah",
				TargetDir: "/blah/dist",
				BinPath:   "/blah/dist/lockbox",
				ZipPath:   "/blah/out/blargle.zip",
			},
		},
	}

	for _, c := range cases {
		desc, config, root, want := c.desc, c.config, c.root, c.want
		t.Run(desc, func(t *testing.T) {
			got, err := config.BuildConfig(root)
			if err != nil {
				t.Fatal(err)
			}
			assertEqual(t, got, want)
		})
	}
}

func assertEqual(t *testing.T, got, want interface{}) {
	diff := cmp.Diff(got, want)
	if diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}
