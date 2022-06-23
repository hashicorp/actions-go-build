package config

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestInputs_Config_ok(t *testing.T) {
	cases := []struct {
		description string
		inputs      Inputs
		rc          RepoContext
		want        Config
	}{
		{
			"standard inputs",
			standardInputs(),
			standardRepoContext(),
			standardConfig(),
		},
		{
			"version +ent",
			testInputs(func(i *Inputs) { i.ProductVersion = "1.2.3+ent" }),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.ProductVersion = "1.2.3+ent"
				c.ZipName = "lockbox_1.2.3+ent_linux_amd64.zip"
				c.PrimaryBuild.ZipPath = "/some/dir/work/out/lockbox_1.2.3+ent_linux_amd64.zip"
				c.VerificationBuild.ZipPath = "/some/dir/verification/out/lockbox_1.2.3+ent_linux_amd64.zip"
			}),
		},
		{
			"version +ent.hsm",
			testInputs(func(i *Inputs) { i.ProductVersion = "1.2.3+ent.hsm" }),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.ProductVersion = "1.2.3+ent.hsm"
				c.ZipName = "lockbox_1.2.3+ent.hsm_linux_amd64.zip"
				c.PrimaryBuild.ZipPath = "/some/dir/work/out/lockbox_1.2.3+ent.hsm_linux_amd64.zip"
				c.VerificationBuild.ZipPath = "/some/dir/verification/out/lockbox_1.2.3+ent.hsm_linux_amd64.zip"
			}),
		},
		{
			"overridden zip_name",
			testInputs(func(i *Inputs) {
				i.ZipName = "blarglefish.zip"
			}),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.ZipName = "blarglefish.zip"
				c.PrimaryBuild.ZipPath = "/some/dir/work/out/blarglefish.zip"
				c.VerificationBuild.ZipPath = "/some/dir/verification/out/blarglefish.zip"
			}),
		},
		{
			"overridden bin_name",
			testInputs(func(i *Inputs) {
				i.BinName = "blarglefish"
			}),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.BinName = "blarglefish"
				c.PrimaryBuild.BinPath = "/some/dir/work/dist/blarglefish"
				c.VerificationBuild.BinPath = "/some/dir/verification/dist/blarglefish"
			}),
		},
		{
			"overridden primary build root",
			testInputs(func(i *Inputs) {
				i.PrimaryBuildRoot = "/other/dir/work"
			}),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.PrimaryBuildRoot = "/other/dir/work"
				c.VerificationBuildRoot = "/other/dir/verification"
				c.PrimaryBuild.BinPath = "/other/dir/work/dist/lockbox"
				c.PrimaryBuild.ZipPath = "/other/dir/work/out/lockbox_1.2.3_linux_amd64.zip"
				c.VerificationBuild.BinPath = "/other/dir/verification/dist/lockbox"
				c.VerificationBuild.ZipPath = "/other/dir/verification/out/lockbox_1.2.3_linux_amd64.zip"
			}),
		},
	}

	for _, c := range cases {
		description, inputs, rc, want := c.description, c.inputs, c.rc, c.want
		t.Run(description, func(t *testing.T) {
			got, err := inputs.Config(rc)
			if err != nil {
				t.Fatal(err)
			}
			diff := deep.Equal(got, want)
			for _, d := range diff {
				t.Error(d)
			}
		})
	}
}

func TestInputs_Config_err(t *testing.T) {
	cases := []struct {
		description string
		inputsSet   []Inputs
		rc          RepoContext
		wantErr     string
	}{
		{
			"empty os",
			[]Inputs{
				testInputs(func(i *Inputs) { i.OS = "" }),
				testInputs(func(i *Inputs) { i.OS = "    " }),
			},
			testRepoContext(),
			"required input 'os' is empty",
		},
		{
			"empty arch",
			[]Inputs{
				testInputs(func(i *Inputs) { i.Arch = "" }),
				testInputs(func(i *Inputs) { i.Arch = "    " }),
			},
			testRepoContext(),
			"required input 'arch' is empty",
		},
		{
			"empty reproducible",
			[]Inputs{
				testInputs(func(i *Inputs) { i.Reproducible = "" }),
				testInputs(func(i *Inputs) { i.Reproducible = "    " }),
			},
			testRepoContext(),
			"required input 'reproducible' is empty",
		},
		{
			"empty instructions",
			[]Inputs{
				testInputs(func(i *Inputs) { i.Instructions = "" }),
				testInputs(func(i *Inputs) { i.Instructions = "    " }),
			},
			testRepoContext(),
			"required input 'instructions' is empty",
		},
	}

	for _, c := range cases {
		description, inputsSet, rc, wantErr := c.description, c.inputsSet, c.rc, c.wantErr
		t.Run(description, func(t *testing.T) {
			for _, inputs := range inputsSet {
				inputs := inputs
				t.Run("", func(t *testing.T) {
					wantDesc := fmt.Sprintf("want error containing %q", wantErr)
					_, err := inputs.Config(rc)
					if err == nil {
						t.Fatalf("got nil error; %s", wantDesc)
					}
					gotErr := fmt.Sprint(err)
					if !strings.Contains(gotErr, wantErr) {
						t.Errorf("got error %q; %s", gotErr, wantDesc)
					}
				})
			}
		})
	}

}

// testInputs generates an Inputs for testing by taking the standard inputs
// and applying the provided modifier functions to it in the order provided.
func testInputs(modifiers ...func(*Inputs)) Inputs {
	return applyModifiers(standardInputs(), modifiers...)
}

// testRepoContext generates a RepoContext for testing by taking the standard inputs
// and applying the provided modifier functions to it in the order provided.
func testRepoContext(modifiers ...func(*RepoContext)) RepoContext {
	return applyModifiers(standardRepoContext(), modifiers...)
}

// testConfig generates a Config for testing by taking the standard inputs
// and applying the provided modifier functions to it in the order provided.
func testConfig(modifiers ...func(*Config)) Config {
	return applyModifiers(standardConfig(), modifiers...)
}

func applyModifiers[T any](to T, modifiers ...func(thing *T)) T {
	for _, m := range modifiers {
		m(&to)
	}
	return to
}

func standardInputs() Inputs {
	return Inputs{
		ProductRepository: "dadgarcorp/lockbox",
		ProductName:       "lockbox",
		ProductVersion:    "1.2.3",
		GoVersion:         "1.18",
		OS:                "linux",
		Arch:              "amd64",
		Reproducible:      "assert",
		Instructions:      "go build -o $BIN_PATH",
		// These are intentionally left blank.
		BinName:               "",
		ZipName:               "",
		PrimaryBuildRoot:      "",
		VerificationBuildRoot: "",
	}
}

func standardRepoContext() RepoContext {
	return RepoContext{
		WorkDir:    "/some/dir/work",
		CommitSHA:  "cabba9e",
		CommitTime: time.Date(2001, time.December, 1, 0, 0, 0, 0, time.Local),
	}
}

func standardConfig() Config {
	return Config{
		Inputs: Inputs{
			ProductRepository:     "dadgarcorp/lockbox",
			ProductName:           "lockbox",
			ProductVersion:        "1.2.3",
			GoVersion:             "1.18",
			OS:                    "linux",
			Arch:                  "amd64",
			Reproducible:          "assert",
			Instructions:          "go build -o $BIN_PATH",
			BinName:               "lockbox",
			ZipName:               "lockbox_1.2.3_linux_amd64.zip",
			PrimaryBuildRoot:      "/some/dir/work",
			VerificationBuildRoot: "/some/dir/verification",
		},
		PrimaryBuild: BuildConfig{
			TargetDir: "dist",
			ZipDir:    "out",
			MetaDir:   "meta",
			BinPath:   "/some/dir/work/dist/lockbox",
			ZipPath:   "/some/dir/work/out/lockbox_1.2.3_linux_amd64.zip",
		},
		VerificationBuild: BuildConfig{
			TargetDir: "dist",
			ZipDir:    "out",
			MetaDir:   "meta",
			BinPath:   "/some/dir/verification/dist/lockbox",
			ZipPath:   "/some/dir/verification/out/lockbox_1.2.3_linux_amd64.zip",
		},
	}
}
