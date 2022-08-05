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

// buildFlags are flags you can pass to any build, be it primary or verification.
type buildFlags struct {
	rebuild bool
	verbose bool
}

func (flags *buildFlags) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&flags.rebuild, "rebuild", false, "re-run the build even if cached")
	fs.BoolVar(&flags.verbose, "v", false, "verbose logging")
}

func (flags *buildFlags) managerOpts() []build.ManagerOption {
	return []build.ManagerOption{
		build.WithLogFunc(flags.logFunc()),
		build.WithForceRebuild(flags.rebuild),
	}
}

func (flags *buildFlags) logFunc() log.Func {
	if flags.verbose {
		return log.Info
	}
	return log.Verbose
}

func (flags *buildFlags) newManager(b build.Build) *build.Manager {
	r := build.NewRunner(b, flags.logFunc())
	return build.NewManager(r, flags.managerOpts()...)
}

func (flags *buildFlags) buildOpts() []build.Option {
	return []build.Option{build.WithLogfunc(flags.logFunc())}
}

// buildOpts has all the buildFlags plus a build config read from the environment.
type buildOpts struct {
	buildFlags
	config config.Config
}

func (opts *buildOpts) ReadEnv() error {
	var err error
	opts.config, err = config.FromEnvironment(tool)
	return err
}

func (opts *buildOpts) Flags(fs *flag.FlagSet) { opts.buildFlags.Flags(fs) }

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
