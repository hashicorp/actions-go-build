package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/internal/config"
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
