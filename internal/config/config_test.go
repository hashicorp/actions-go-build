package config

import (
	"testing"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/assert"
	"github.com/hashicorp/go-version"
)

func TestConfig_init_ok(t *testing.T) {
	cases := []struct {
		description string
		inputs      Config
		rc          crt.RepoContext
		want        Config
	}{
		{
			"standard inputs",
			standardUnintializedConfig(),
			standardRepoContext(),
			standardConfig(),
		},
		{
			"version +ent",
			testUninitializedConfig(func(i *Config) {
				i.Product.Version.Full = "1.2.3+ent"
			}),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.Product.Version.Full = "1.2.3+ent"
				c.Product.Version.Core = "1.2.3"
				c.Product.Version.Meta = "ent"
				c.ZipName = "lockbox_1.2.3+ent_linux_amd64.zip"
			}),
		},
		{
			"version +ent.hsm",
			testUninitializedConfig(func(i *Config) {
				i.Product.Version.Full = "1.2.3+ent.hsm"
			}),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.Product.Version.Full = "1.2.3+ent.hsm"
				c.Product.Version.Core = "1.2.3"
				c.Product.Version.Meta = "ent.hsm"
				c.ZipName = "lockbox_1.2.3+ent.hsm_linux_amd64.zip"
			}),
		},
		{
			"overridden zip_name",
			testUninitializedConfig(func(i *Config) {
				i.ZipName = "blarglefish.zip"
			}),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.ZipName = "blarglefish.zip"
			}),
		},
		{
			"overridden bin_name",
			testUninitializedConfig(func(i *Config) {
				i.Product.ExecutableName = "blarglefish"
			}),
			testRepoContext(),
			testConfig(func(c *Config) {
				c.Product.ExecutableName = "blarglefish"
			}),
		},
		{
			"overridden primary build root",
			testUninitializedConfig(func(i *Config) {
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
			got, err := inputs.init(rc)
			if err != nil {
				t.Fatal(err)
			}

			// Test the verification build root separately because it's
			// a temp directory with an unpredictable name.
			if got.VerificationBuildRoot == "" {
				t.Errorf("got empty VerificationBuildRoot")
			}
			// Force got and want to be empty so we can assert equality on
			// everything else.
			got.VerificationBuildRoot = ""
			want.VerificationBuildRoot = ""

			assert.Equal(t, got, want)
		})
	}
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

func testUninitializedConfig(modifiers ...func(*Config)) Config {
	return applyModifiers(standardUnintializedConfig(), modifiers...)
}

func applyModifiers[T any](to T, modifiers ...func(thing *T)) T {
	for _, m := range modifiers {
		m(&to)
	}
	return to
}

func standardProduct() crt.Product {
	return crt.Product{
		Repository:     "dadgarcorp/lockbox",
		Name:           "lockbox",
		CoreName:       "lockbox",
		ExecutableName: "lockbox",
		Version: crt.ProductVersion{
			Full: "1.2.3",
			Core: "1.2.3",
			Meta: "",
		},
		Revision:     "cabba9e",
		RevisionTime: standardCommitTimeRFC3339(),
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
		RepoName:    "dadgarcorp/lockbox",
		CoreVersion: *version.Must(version.NewVersion("1.2.3")),
		RootDir:     "/some/dir/work",
		Dir:         "/some/dir/work",
		CommitSHA:   "cabba9e",
		CommitTime:  standardCommitTimestamp(),
	}
}

func standardConfig() Config {
	return Config{
		Product:               standardProduct(),
		Parameters:            standardParameters(),
		Reproducible:          "assert",
		PrimaryBuildRoot:      "/some/dir/work",
		VerificationBuildRoot: "/some/dir/verification",
		ZipName:               "lockbox_1.2.3_linux_amd64.zip",
	}
}

func standardUnintializedConfig() Config {
	return Config{
		Product: crt.Product{
			Repository:     "",
			Directory:      "",
			Name:           "",
			CoreName:       "",
			ExecutableName: "",
			Version: crt.ProductVersion{
				Full: "",
				Core: "",
				Meta: "",
			},
			Revision:     "",
			RevisionTime: "",
		},
		Parameters: crt.BuildParameters{
			GoVersion:    "1.18",
			Instructions: "",
			OS:           "linux",
			Arch:         "amd64",
		},
		Reproducible:          "",
		ZipName:               "",
		PrimaryBuildRoot:      "",
		VerificationBuildRoot: "",
	}
}

func standardParameters() crt.BuildParameters {
	return crt.BuildParameters{
		GoVersion:    "1.18",
		OS:           "linux",
		Arch:         "amd64",
		Instructions: `go build -o "$BIN_PATH" -trimpath -buildvcs=false`,
	}
}
