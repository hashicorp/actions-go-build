package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/config"
	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/build"
)

const buildInstructionsHelp = `
You can see the build instructions by running the 'config action' subcommand.
The instructions are set by the BUILD_INSTRUCTIONS environment variable, or a
simple default set of build instructions are used if that is not set.

This command fails if the build instructions do not write a file to BIN_PATH.

See the 'config env describe' subcommand for info on what environment variables are
available to your build instructions.

See the 'config env dump' subcommand to print out the values for all these variables.
`

type buildOpts struct {
	rebuild bool
	verbose bool
	config  config.Config
}

func (opts *buildOpts) ReadEnv() error {
	var err error
	opts.config, err = config.FromEnvironment()
	return err
}

func (opts *buildOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&opts.rebuild, "rebuild", false, "re-run the build even if cached")
	fs.BoolVar(&opts.verbose, "v", false, "verbose logging")
}

func (opts *buildOpts) primaryBuild() (*build.Manager, error) {
	pc, err := opts.config.PrimaryBuildConfig()
	if err != nil {
		return nil, err
	}
	b, err := build.NewPrimary(pc, opts.buildOpts()...)
	if err != nil {
		return nil, err
	}
	return opts.newManager(b), nil
}

func (opts *buildOpts) managerOpts() []build.ManagerOption {
	return []build.ManagerOption{
		build.WithLogFunc(opts.logFunc()),
		build.WithForceRebuild(opts.rebuild),
	}
}

func (opts *buildOpts) newManager(b build.Build) *build.Manager {
	r := build.NewRunner(b, opts.logFunc())
	return build.NewManager(r, opts.managerOpts()...)
}

func (opts *buildOpts) buildOpts() []build.Option {
	return []build.Option{build.WithLogfunc(opts.logFunc())}
}

func (opts *buildOpts) logFunc() log.Func {
	if opts.verbose {
		return log.Info
	}
	return log.Verbose
}
