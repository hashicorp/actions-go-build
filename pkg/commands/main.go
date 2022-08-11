package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// stdout and stderr can be overridden for testing.
var stdout = os.Stdout
var stderr = os.Stderr

// Main is the root command of the whole CLI. It is given the name "go" so that
// when this CLI is incorporated into a parent CLI, the commands within will be
// rooted at "go". E.g. "go-build", "go-build primary", "go-build verification".
var Main = cli.RootCommand("go-build", "go build and related functions",
	Build, Verify, Config)

type buildOpts struct {
	buildish
	verification bool
}

func (opts *buildOpts) Flags(fs *flag.FlagSet) {
	opts.buildish.Flags(fs)
	fs.BoolVar(&opts.verification, "verification", false, "configure build as a verification build")
}

var Build = cli.LeafCommand("build", "run a build", func(opts *buildOpts) error {
	return opts.runBuild("Running build", opts.verification)
})

var Describe = cli.RootCommand("describe", "describe things", DescribeBuildEnv)

var DescribeBuildEnv = cli.LeafCommand("build-env", "describe the build environment variables", func(cli.None) error {
	env := build.BuildEnvDefinitions()
	return cli.TabWrite(stdout, env, func(e build.EnvVar) string {
		return fmt.Sprintf("%s\t%s", e.Name, e.Description)
	})
})

func printList(list []string) error {
	return cli.TabWrite(stdout, list, func(s string) string { return s })
}

var Verify = cli.LeafCommand("verify", "verify a build's reproducibility", func(v *verifyish) error {
	return v.runVerification()
})
