package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	"github.com/hashicorp/composite-action-framework-go/pkg/github"
)

type verifyOpts struct {
	Builds       opts.AllBuilds
	ActionConfig opts.ActionConfig
	GitHub       opts.GitHubOpts
	StepSummary  github.StepSummary
	ResultWriter opts.ResultWriter

	// internal opts used for different flavours of verification.
	noRunPrimaryBuild, noRunVerificationBuild bool
}

func (bo *verifyOpts) ReadEnv() error {
	return cli.ReadEnvAll(&bo.Builds, &bo.ActionConfig, &bo.GitHub, &bo.StepSummary)
}

func (bo *verifyOpts) Flags(fs *flag.FlagSet) { cli.FlagsAll(fs, &bo.GitHub, &bo.StepSummary) }

var BuildAndVerify = cli.LeafCommand("build-and-verify", "run primary and verification builds; assert match", func(opts *verifyOpts) error {
	return verifyCore(opts)
}).WithHelp(`
Run the primary and verification builds and verify that their outputs are identical.

This command is mostly useful for running locally, when you want to ensure you haven't
introduced accidental nondeterminism into the build. It caches builds, so running it
twice without making any changes to the source code will just use the cached results
on the second and subsequent runs.

Build results are cached according to the SourceHash of the current working directory.
If the working directory is clean (no new or modified files) then the SourceHash is the
same as the HEAD commit SHA. Otherwise it's a hash of that SHA plus the contents of all
new and changed files.
`)
