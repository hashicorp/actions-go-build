package config

import (
	"testing"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/assert"
)

func TestInputs_Config_ok(t *testing.T) {
	cases := []struct {
		description string
		inputs      Inputs
		rc          crt.RepoContext
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
			testInputs(func(i *Inputs) { i.Product.Version = "1.2.3+ent" }),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.Product.Version = "1.2.3+ent"
				c.ZipName = "lockbox_1.2.3+ent_linux_amd64.zip"
			}),
		},
		{
			"version +ent.hsm",
			testInputs(func(i *Inputs) { i.Product.Version = "1.2.3+ent.hsm" }),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.Product.Version = "1.2.3+ent.hsm"
				c.ZipName = "lockbox_1.2.3+ent.hsm_linux_amd64.zip"
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
			assert.Equal(t, got, want)
		})
	}
}

// testInputs generates an Inputs for testing by taking the standard inputs
// and applying the provided modifier functions to it in the order provided.
func testInputs(modifiers ...func(*Inputs)) Inputs {
	return applyModifiers(standardInputs(), modifiers...)
}

// testRepoContext generates a RepoContext for testing by taking the standard
// RepoContext and applying the provided modifier functions to it in the order
// provided.
func testRepoContext(modifiers ...func(*crt.RepoContext)) crt.RepoContext {
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

func standardProduct() crt.Product {
	return crt.Product{
		Repository:   "github.com/dadgarcorp/lockbox",
		Name:         "lockbox",
		CoreName:     "lockbox",
		Version:      "1.2.3",
		Revision:     "cabba9e",
		RevisionTime: standardCommitTimeRFC3339(),
	}
}

func standardInputs() Inputs {
	return Inputs{
		Product:      standardProduct(),
		GoVersion:    "1.18",
		OS:           "linux",
		Arch:         "amd64",
		Reproducible: "assert",
		Instructions: `go build -o "$BIN_PATH"`,
		// These are intentionally left blank.
		BinName:               "",
		ZipName:               "",
		PrimaryBuildRoot:      "",
		VerificationBuildRoot: "",
	}
}

func standardCommitTimestamp() time.Time {
	return time.Date(2022, time.June, 30, 10, 31, 6, 0, time.UTC)
}

func standardCommitTimeRFC3339() string {
	return "2022-06-30T10:31:06Z"
}

func standardRepoContext() crt.RepoContext {
	return crt.RepoContext{
		Dir:        "/some/dir/work",
		CommitSHA:  "cabba9e",
		CommitTime: standardCommitTimestamp(),
	}
}

func standardConfig() Config {
	return Config{
		Inputs: Inputs{
			Product:               standardProduct(),
			GoVersion:             "1.18",
			OS:                    "linux",
			Arch:                  "amd64",
			Reproducible:          "assert",
			Instructions:          `go build -o "$BIN_PATH"`,
			MainPackage:           ".",
			BinName:               "lockbox",
			ZipName:               "lockbox_1.2.3_linux_amd64.zip",
			PrimaryBuildRoot:      "/some/dir/work",
			VerificationBuildRoot: "/some/dir/verification",
		},
		TargetDir: "dist",
		ZipDir:    "out",
		MetaDir:   "meta",
	}
}
