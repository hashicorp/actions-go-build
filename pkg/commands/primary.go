package commands

import (
	"os"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

const primaryHelp = `
Run the primary build by executing the build instructions in the current directory.
`

// BuildPrimary runs the primary build, in the current directory.
var BuildPrimary = cli.LeafCommand("primary", "run the primary build", func(opts *pBuildOpts) error {

	pb, err := opts.primaryBuild()
	if err != nil {
		return err
	}

	result, err := pb.Result()
	if err != nil {
		return err
	}
	if opts.verbose {
		if err := dumpJSON(os.Stdout, result); err != nil {
			return err
		}
	}

	return result.Error()
}).WithHelp(primaryHelp + buildInstructionsHelp)

type pBuildOpts struct {
	buildOpts
}

func (opts *pBuildOpts) build() (*build.Manager, error) {
	return opts.primaryBuild()
}
