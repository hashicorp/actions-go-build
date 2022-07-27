package commands

import (
	"log"

	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// Primary runs the primary build, in the current directory.
var Primary = cli.LeafCommand("primary", "run the primary build", func(b *opts.PrimaryBuild) error {
	result := b.Build.Run()

	resultFile, err := b.ResultWriter.WriteBuildResult(result)
	if err != nil {
		return err
	}
	if resultFile != "" {
		log.Printf("Primary build results written to %q", resultFile)
	}

	if err := cacheResult("Primary", result); err != nil {
		return err
	}

	return result.Error()
})
