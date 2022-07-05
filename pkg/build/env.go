package build

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/pkg/crt"
)

// envVar represents a documented environment variable alongside
// a function showing how it is extracted from crt.BuildConfig.
type envVar struct {
	name, desc string
	valueFunc  func(crt.BuildConfig) string
}

// String emits an env var string compatible with exec.CMD.Env.
func (ev envVar) String(c crt.BuildConfig) string {
	return fmt.Sprintf("%s=%s", ev.name, ev.valueFunc(c))
}

// EnvVar represents a documented environment variable.
type EnvVar struct {
	Name, Description string
}

// BuildEnvDefinitions returns the set of env vars guaranteed
// to be available to the build instructions, alongside a description
// of each one.
func BuildEnvDefinitions() []EnvVar {
	bed := buildEnvDef()
	out := make([]EnvVar, len(bed))
	for i, e := range bed {
		out[i] = EnvVar{Name: e.name, Description: e.desc}
	}
	return out
}

func buildEnvDef() []envVar {
	return []envVar{
		{
			"TARGET_DIR",
			"Absolute path to the zip contents directory.",
			func(c crt.BuildConfig) string { return c.TargetDir },
		},
		{
			"PRODUCT_NAME",
			"Same as the `product_name` input.",
			func(c crt.BuildConfig) string { return c.Product.Name },
		},
		{
			"PRODUCT_VERSION",
			"Same as the `product_version` input.",
			func(c crt.BuildConfig) string { return c.Product.Version },
		},
		{
			"PRODUCT_REVISION",
			"The git commit SHA of the product repo being built.",
			func(c crt.BuildConfig) string { return c.Product.Revision },
		},
		{
			"PRODUCT_REVISION_TIME",
			"UTC timestamp of the `PRODUCT_REVISION` commit in iso-8601 format.",
			func(c crt.BuildConfig) string { return c.Product.RevisionTime },
		},
		// NOTE omitting BIN_NAME as not strictly needed.
		{
			"BIN_PATH",
			"Absolute path to where instructions must write Go executable.",
			func(c crt.BuildConfig) string { return c.BinPath },
		},
		{
			"OS",
			"Same as the `os` input.",
			func(c crt.BuildConfig) string { return c.TargetOS },
		},
		{
			"ARCH",
			"Same as the `arch` input.",
			func(c crt.BuildConfig) string { return c.TargetArch },
		},
		{
			"GOOS",
			"Same as `OS`.",
			func(c crt.BuildConfig) string { return c.TargetOS },
		},
		{
			"GOARCH",
			"Same as `ARCH`.",
			func(c crt.BuildConfig) string { return c.TargetArch },
		},
	}
}

// buildEnv materialises the values for each defined env var as a slice
// compatible with exec.CMD.Env.
func (b *build) buildEnv() []string {
	bed := buildEnvDef()
	env := make([]string, len(bed))
	for i, e := range buildEnvDef() {
		env[i] = e.String(b.config)
	}
	return env
}
