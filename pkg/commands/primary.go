package commands

import (
	"fmt"
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

	path, err := result.Save()
	if err != nil {
		return fmt.Errorf("Failed to cache build results: %s", err)
	}
	log.Printf("Build results cached to %s", path)

	return result.Error()
})
