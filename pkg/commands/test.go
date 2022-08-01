package commands

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/verifier"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

var Test = cli.LeafCommand("test", "test reproducibility of current worktree + config", func(opts *testOpts) error {
	verifier := verifier.New(opts.pBuild, opts.vBuild)
	result, err := verifier.Verify()
	if err != nil {
		return err
	}
	if opts.verbose {
		if err := dumpJSON(os.Stdout, result); err != nil {
			return err
		}
	}
	return result.Error()

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

func dumpJSON(w io.Writer, v any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", " ")
	return e.Encode(v)
}

type testOpts struct {
	rebuildAll bool
	verbose    bool

	pOpts pBuildOpts
	vOpts vBuildOpts

	pBuild, vBuild *build.Manager
}

func (opts *testOpts) ReadEnv() error {
	return cli.ReadEnvAll(&opts.pOpts, &opts.vOpts)
}

func (opts *testOpts) Flags(fs *flag.FlagSet) {
	opts.vOpts.ownFlags(fs)
	fs.BoolVar(&opts.verbose, "v", false, "verbose logging")
	fs.BoolVar(&opts.pOpts.rebuild, "rebuild-p", false, "re-run the primary build even if cached")
	fs.BoolVar(&opts.vOpts.rebuild, "rebuild-v", false, "re-run the verification build even if cached")
	fs.BoolVar(&opts.rebuildAll, "rebuild", false, "re-run both builds, ignoring the cache")
}

func (opts *testOpts) Init() error {
	opts.pOpts.verbose = opts.verbose
	opts.vOpts.verbose = opts.verbose

	opts.pOpts.rebuild = opts.pOpts.rebuild || opts.rebuildAll
	opts.vOpts.rebuild = opts.vOpts.rebuild || opts.rebuildAll

	var err error
	if opts.pBuild, err = opts.pOpts.primaryBuild(); err != nil {
		return err
	}
	if opts.vBuild, err = opts.vOpts.verificationBuild(); err != nil {
		return err
	}
	return nil
}
