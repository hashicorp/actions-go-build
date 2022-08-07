package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

const testHelp = `
Run the primary and verification builds and verify that their outputs are identical.

This command is mostly useful for running locally, when you want to ensure you haven't
introduced accidental nondeterminism into the build. It caches builds, so running it
twice without making any changes to the source code will just use the cached results
on the second and subsequent runs.

Build results are cached according to the SourceHash of the current working directory.
If the working directory is clean (no new or modified files) then the SourceHash is the
same as the HEAD commit SHA. Otherwise it's a hash of that SHA plus the contents of all
new and changed files.
`

type testOpts struct {
	logOpts
	rebuildAll     bool
	output         outputOpts
	primary        pbOpts
	verification   lvbOpts
	pBuild, vBuild *build.Manager
}

func (opts *testOpts) ReadEnv() error {
	return cli.ReadEnvAll(&opts.output, &opts.primary, &opts.verification)
}

func (opts *testOpts) Flags(fs *flag.FlagSet) {
	cli.FlagFuncsAll(fs, opts.logOpts.Flags, opts.output.ownFlags, opts.primary.ownFlags, opts.verification.ownFlags)
	fs.BoolVar(&opts.primary.rebuild, "rebuild-p", false, "re-run the primary build even if cached")
	fs.BoolVar(&opts.verification.rebuild, "rebuild-v", false, "re-run the verification build even if cached")
	fs.BoolVar(&opts.rebuildAll, "rebuild", false, "re-run both builds, ignoring the cache")
}

func (opts *testOpts) Init() error {
	opts.output.logOpts = opts.logOpts

	opts.primary.rebuild = opts.primary.rebuild || opts.rebuildAll
	opts.primary.logOpts = opts.logOpts
	opts.primary.output.logOpts = opts.logOpts

	opts.verification.rebuild = opts.verification.rebuild || opts.rebuildAll
	opts.verification.logOpts = opts.logOpts
	opts.verification.output.logOpts = opts.logOpts
	opts.verification.primary.logOpts = opts.logOpts
	opts.verification.primary.output.logOpts = opts.logOpts

	if err := cli.InitAll(&opts.primary, &opts.verification); err != nil {
		return err
	}

	var err error
	if opts.pBuild, err = opts.primary.build(); err != nil {
		return err
	}
	opts.vBuild, err = opts.verification.build()
	return err
}

var Test = cli.LeafCommand("test", "test reproducibility of current worktree + config", func(opts *testOpts) error {
	verifier, err := opts.newVerifier(opts.pBuild, opts.vBuild)
	if err != nil {
		return err
	}
	result, err := verifier.Verify()
	if err != nil {
		return err
	}
	return opts.output.result("Reproducibility test", result)
}).WithHelp(testHelp)
