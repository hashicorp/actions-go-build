package commands

import (
	"flag"
	"os"

	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// stdout and stderr can be overridden for testing.
var stdout = os.Stdout
var stderr = os.Stderr

// Main is the root command of the whole CLI. It is given the name "go" so that
// when this CLI is incorporated into a parent CLI, the commands within will be
// rooted at "go". E.g. "go-build", "go-build primary", "go-build verification".
var Main = cli.RootCommand("go-build", "go build and related functions",
	Build, Verify2, BuildEnv, Config)

type buildOpts struct {
	buildish
	verification bool
}

func (opts *buildOpts) Flags(fs *flag.FlagSet) {
	opts.buildish.Flags(fs)
	fs.BoolVar(&opts.verification, "verification", false, "force a verification build")
}

var Build = cli.LeafCommand("build", "run a build", func(b *buildOpts) error {
	return b.runBuild(b.verification)
})

var Verify2 = cli.LeafCommand("verify", "verify a build's reproducibility", func(v *verifyish) error {
	return v.runVerification()
})

//var Build = cli.RootCommand("build", "run builds and inspect the build env",
//	PrimaryBuild, LVBuild, BuildEnv)

var BuildEnv = cli.RootCommand("build-env", "inspect the build environment",
	BuildEnvDescribe, BuildEnvDump)
