package build

import (
	"fmt"
)

// envVar represents a documented environment variable alongside
// a function showing how it is extracted from BuildConfig.
type EnvVar struct {
	Name, Description string
	valueFunc         func(BuildConfig) string
}

// Env materialises the values for each defined env var as a slice
// compatible with exec.CMD.Env.
func (b *build) Env() []string {
	bed := BuildEnvDefinitions()
	env := make([]string, len(bed))
	for i, e := range bed {
		env[i] = fmt.Sprintf("%s=%s", e.Name, e.valueFunc(b.config))
	}
	return env
}

// BuildEnvDefinitions returns the set of env vars guaranteed
// to be available to the build instructions, alongside a description
// of each one.
func BuildEnvDefinitions() []EnvVar {
	return []EnvVar{
		{
			"TARGET_DIR",
			"Absolute path to the zip contents directory.",
			func(c BuildConfig) string { return c.Paths.TargetDir },
		},
		{
			"PRODUCT_NAME",
			"Same as the `product_name` input.",
			func(c BuildConfig) string { return c.Product.Name },
		},
		{
			"PRODUCT_VERSION",
			"Same as the `product_version` input.",
			func(c BuildConfig) string { return c.Product.Version.Full },
		},
		{
			"PRODUCT_REVISION",
			"The git commit SHA of the product repo being built.",
			func(c BuildConfig) string { return c.Product.Revision },
		},
		{
			"PRODUCT_REVISION_TIME",
			"UTC timestamp of the `PRODUCT_REVISION` commit in iso-8601 format.",
			func(c BuildConfig) string { return c.Product.RevisionTime },
		},
		{
			"BIN_PATH",
			"Absolute path to where instructions must write Go executable.",
			func(c BuildConfig) string { return c.Paths.BinPath },
		},
		{
			"OS",
			"Same as the `os` input.",
			func(c BuildConfig) string { return c.Parameters.OS },
		},
		{
			"ARCH",
			"Same as the `arch` input.",
			func(c BuildConfig) string { return c.Parameters.Arch },
		},
		{
			"GOOS",
			"Same as `OS`.",
			func(c BuildConfig) string { return c.Parameters.OS },
		},
		{
			"GOARCH",
			"Same as `ARCH`.",
			func(c BuildConfig) string { return c.Parameters.Arch },
		},
		{
			"WORKTREE_DIRTY",
			"Whether the workrtree is dirty (`true` or `false`).",
			func(c BuildConfig) string { return fmt.Sprint(c.Product.IsDirty()) },
		},
		{
			"WORKTREE_HASH",
			"Unique hash of the work tree. Same as PRODUCT_REVISION unless WORKTREE_DIRTY.",
			func(c BuildConfig) string { return fmt.Sprint(c.Product.SourceHash) },
		},
	}
}
