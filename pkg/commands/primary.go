package commands

import (
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// Primary runs the primary build, in the current directory.
var Primary = cli.LeafCommand("primary", "run the primary build", func(b *opts.PrimaryBuild) error {
	return b.Run()
})
