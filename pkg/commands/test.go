package commands

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/verifier"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	cp "github.com/otiai10/copy"
)

type testOpts struct {
	rebuildP   bool
	rebuildV   bool
	rebuildAll bool
	v          bool
	config     config.Config

	pBuild, vBuild *build.Manager
}

func (opts *testOpts) ReadEnv() error {
	var err error
	opts.config, err = config.FromEnvironment()
	return err
}

func (opts *testOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.v, "v", false, "verbose logging")
	fs.BoolVar(&opts.rebuildP, "rebuild-p", false, "re-run the primary build even if cached")
	fs.BoolVar(&opts.rebuildV, "rebuild-v", false, "re-run the verification build even if cached")
	fs.BoolVar(&opts.rebuildAll, "rebuild", false, "re-run both builds, ignoring the cache")
}

func (opts *testOpts) Init() error {
	var err error
	if opts.pBuild, err = opts.primaryBuild(); err != nil {
		return err
	}
	if opts.vBuild, err = opts.verificationBuild(); err != nil {
		return err
	}
	return nil
}

var Test = cli.LeafCommand("test", "test reproducibility of current worktree + config", func(opts *testOpts) error {

	verifier := verifier.New(opts.pBuild, opts.vBuild)
	result, err := verifier.Verify()
	if err != nil {
		return err
	}
	if opts.v {
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

func (opts *testOpts) primaryBuild() (*build.Manager, error) {
	pc, err := opts.config.PrimaryBuildConfig()
	if err != nil {
		return nil, err
	}
	b, err := opts.build(pc, opts.rebuildAll || opts.rebuildP)
	if err != nil {
		return nil, err
	}
	return opts.newManager(b), nil
}

func (opts *testOpts) verificationBuild() (*build.Manager, error) {
	vc, err := opts.config.VerificationBuildConfig()
	if err != nil {
		return nil, err
	}

	b, err := opts.build(vc, opts.rebuildAll || opts.rebuildV)
	if err != nil {
		return nil, err
	}
	pc, err := opts.config.PrimaryBuildConfig()
	if err != nil {
		return nil, err
	}
	preBuild := func(build.Build) error {
		pPath := pc.Paths.WorkDir
		vPath := vc.Paths.WorkDir
		return cp.Copy(pPath, vPath)
	}
	return opts.newManager(b, build.WithPreBuild(preBuild)), nil
}

func (opts *testOpts) newManager(b build.Build, options ...build.ManagerOption) *build.Manager {
	options = append(options, build.WithLogFunc(opts.logFunc()))
	return build.NewManager(b, options...)
}

func (opts *testOpts) logFunc() log.Func {
	if opts.v {
		return log.Info
	}
	return log.Verbose
}

func (opts *testOpts) build(c build.Config, rebuild bool) (build.Build, error) {
	return build.New(c, build.WithLogfunc(opts.logFunc()))
}
