package commands

import (
	"flag"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

const primaryHelp = `
Run the primary build by executing the build instructions in the current directory.
` + buildInstructionsHelp

type pBuildOpts struct {
	present presenter
	buildOpts
}

func (opts *pBuildOpts) ReadEnv() error         { return cli.ReadEnvAll(&opts.present, &opts.buildOpts) }
func (opts *pBuildOpts) Flags(fs *flag.FlagSet) { cli.FlagsAll(fs, &opts.present, &opts.buildOpts) }

// BuildPrimary runs the primary build, in the current directory.
var BuildPrimary = cli.LeafCommand("primary", "run the primary build", func(opts *pBuildOpts) error {
	pb, err := opts.build()
	if err != nil {
		return err
	}

	result, err := pb.Result()
	if err != nil {
		return err
	}

	return opts.present.result("Primary build", result)

}).WithHelp(primaryHelp)

func (opts *pBuildOpts) build() (*build.Manager, error) { return opts.primaryBuild() }
