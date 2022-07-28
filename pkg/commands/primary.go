package commands

import (
	"github.com/hashicorp/actions-go-build/internal/log"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// BuildPrimary runs the primary build, in the current directory.
var BuildPrimary = cli.LeafCommand("primary", "run the primary build", func(b *opts.PrimaryBuild) error {
	result := b.Build.Run()

	resultFile, err := b.ResultWriter.WriteBuildResult(result)
	if err != nil {
		return err
	}
	if resultFile != "" {
		log.Info("Primary build results written to %q", resultFile)
	}

	if err := cacheResult("Primary", result); err != nil {
		return err
	}

	return result.Error()
}).WithHelp(`
Run the primary build by executing the build instructions in the current directory.
` + buildInstructionsHelp)

const buildInstructionsHelp = `
You can see the build instructions by running the 'config action' subcommand.
The instructions are set by the BUILD_INSTRUCTIONS environment variable, or a
simple default set of build instructions are used if that is not set.

This command fails if the build instructions do not write a file to BIN_PATH.

See the 'config env describe' subcommand for info on what environment variables are
available to your build instructions.

See the 'config env dump' subcommand to print out the values for all these variables.
`
